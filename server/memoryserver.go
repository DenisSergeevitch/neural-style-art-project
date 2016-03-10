package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/mgilbir/neural-style-art-project/pb"
	_ "github.com/mgilbir/neural-style-art-project/server/statik"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

type memoryServer struct {
	PendingJobs    map[jobKey]*Job
	InProgressJobs map[jobKey]*Job
	CompletedJobs  map[jobKey]*Job
	Styles         map[string][]byte
	OutputDir      string
	lock           sync.RWMutex
}

func NewMemoryServer(outputDir string) (Server, error) {
	return &memoryServer{
		PendingJobs:    make(map[jobKey]*Job),
		InProgressJobs: make(map[jobKey]*Job),
		CompletedJobs:  make(map[jobKey]*Job),
		Styles:         make(map[string][]byte),
		OutputDir:      outputDir,
	}, nil
}

func (s *memoryServer) LoadStyle(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	style, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	_, name := path.Split(filename)
	ext := len(path.Ext(name))
	s.Styles[name[:len(name)-ext]] = style

	return nil
}

func (s *memoryServer) CreateJob(ctx context.Context, in *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {

	for styleName, style := range s.Styles {
		err := s.createJob(ctx, in.Name, styleName, style, in.Content.Image)
		if err != nil {
			return &pb.CreateJobResponse{}, err
		}
	}

	return &pb.CreateJobResponse{}, nil
}

func (s *memoryServer) CreateFullJob(ctx context.Context, in *pb.CreateFullJobRequest) (*pb.CreateFullJobResponse, error) {
	err := s.createJob(ctx, in.Name, in.Style.Title, in.Style.Image, in.Content.Image)
	return &pb.CreateFullJobResponse{}, err
}

func (s *memoryServer) createJob(ctx context.Context, name string, styleName string, styleImage []byte, contentImage []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	idStr := strings.Replace(id.String(), "-", "", -1)
	key := jobKey{
		ID:        idStr,
		Name:      name,
		Completed: false,
	}

	job := Job{
		Name:           name,
		StyleName:      styleName,
		StyleImage:     styleImage,
		ContentImage:   contentImage,
		PartialResults: make([][]byte, 0),
		LastUpdated:    time.Now(),
	}

	s.PendingJobs[key] = &job
	log.Printf("Added job with id: %q for name: %q and style: %q\n", key.ID, key.Name, styleName)

	//Save the data locally
	jobDir := getDirectory(s.OutputDir, key.ID, key.Name)

	//TODO: cleanup
	styleFilename, err := prepareFilename(jobDir, job.StyleName+".jpg")
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(styleFilename, job.StyleImage, 0755)
	if err != nil {
		log.Println(err)
	}

	contentFilename, err := prepareFilename(jobDir, job.Name+".jpg")
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(contentFilename, job.ContentImage, 0755)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (s *memoryServer) Close() error {
	return nil
}

func (s *memoryServer) RequestJob(ctx context.Context, in *pb.JobRequest) (*pb.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var k jobKey
	var v *Job

	for k, v = range s.PendingJobs {
		//Take 1 item from the map
		break
	}

	if v == nil {
		//No items pending
		return &pb.Job{}, nil
	}

	v.LastUpdated = time.Now()
	s.InProgressJobs[k] = v
	delete(s.PendingJobs, k)

	return &pb.Job{
		Id:   k.ID,
		Name: k.Name,
		Style: &pb.InputImage{
			Title:  v.StyleName,
			Format: pb.ImageFormat_JPG,
			Image:  v.StyleImage,
		},
		Content: &pb.InputImage{
			Title:  v.Name,
			Format: pb.ImageFormat_JPG,
			Image:  v.ContentImage,
		},
	}, nil
}

func (s *memoryServer) AcknowledgeJob(ctx context.Context, in *pb.JobAck) (*pb.JobAck, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return &pb.JobAck{}, fmt.Errorf("Not implemented")
}

func (s *memoryServer) ProgressReport(ctx context.Context, in *pb.JobResult) (*pb.JobProgressResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := jobKey{
		ID:        in.Id,
		Name:      in.Name,
		Completed: false,
	}

	v, ok := s.InProgressJobs[key]
	if !ok {
		return &pb.JobProgressResponse{}, fmt.Errorf("Key with ID %q not found\n", in.Id)
	}

	v.LastUpdated = time.Now()
	v.PartialResults = append(v.PartialResults, in.Image)

	log.Printf("Received progress on id: %q - %q: %d iterations", key.ID, key.Name, in.ProgressCount)

	jobDir := getDirectory(s.OutputDir, key.ID, key.Name)

	//Save to file
	progressFilename, err := prepareFilename(jobDir, fmt.Sprintf("result_%d.png", in.ProgressCount))
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(progressFilename, in.Image, 0755)
	if err != nil {
		log.Println(err)
	}
	return &pb.JobProgressResponse{}, nil
}

func (s *memoryServer) CompleteJob(ctx context.Context, in *pb.JobResult) (*pb.JobResultResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := jobKey{
		ID:        in.Id,
		Name:      in.Name,
		Completed: false,
	}

	v, ok := s.InProgressJobs[key]
	if !ok {
		return &pb.JobResultResponse{}, fmt.Errorf("Key with ID %q not found\n", in.Id)
	}

	v.Result = in.Image
	v.LastUpdated = time.Now()

	delete(s.InProgressJobs, key)
	key.Completed = true
	s.CompletedJobs[key] = v

	log.Printf("Completed id: %q - %q", key.ID, key.Name)

	jobDir := getDirectory(s.OutputDir, key.ID, key.Name)

	//Save to file
	progressFilename, err := prepareFilename(jobDir, "result.png")
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(progressFilename, in.Image, 0755)
	if err != nil {
		log.Println(err)
	}

	return &pb.JobResultResponse{}, nil
}

func (s *memoryServer) FailJob(ctx context.Context, in *pb.JobFail) (*pb.JobFail, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := jobKey{
		ID:        in.Id,
		Name:      in.Name,
		Completed: false,
	}

	v, ok := s.InProgressJobs[key]
	if !ok {
		return &pb.JobFail{}, fmt.Errorf("Key with ID %q not found\n", in.Id)
	}

	v.LastUpdated = time.Now()

	s.PendingJobs[key] = v
	delete(s.InProgressJobs, key)

	log.Printf("Failed id: %q - %q", key.ID, key.Name)

	return &pb.JobFail{}, nil
}

func (s *memoryServer) GetAllJobs(ctx context.Context) (*AllJobsResponse, error) {
	r := AllJobsResponse{}
	for k, v := range s.PendingJobs {
		var progressUrls []string
		for i, _ := range v.PartialResults {
			progressUrls = append(progressUrls, fmt.Sprintf("progress/%s/%s/%d", k.Name, k.ID, i))
		}

		r.Jobs = append(r.Jobs, JobResponse{
			ID:                k.ID,
			Name:              k.Name,
			Status:            "Pending",
			StyleImageUrl:     fmt.Sprintf("style/%s/%s", k.Name, k.ID),
			ContentImageUrl:   fmt.Sprintf("content/%s/%s", k.Name, k.ID),
			ProgressImageUrls: progressUrls,
			ResultImageUrl:    fmt.Sprintf("result/%s/%s", k.Name, k.ID),
		})
	}

	r.Stats = JobStats{
		PendingJobsCount:    len(s.PendingJobs),
		InProgressJobsCount: len(s.InProgressJobs),
		CompletedJobsCount:  len(s.CompletedJobs),
	}

	return &r, nil
}

func (s *memoryServer) GetStyleImage(ctx context.Context, jobId string, name string) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}

func (s *memoryServer) GetContentImage(ctx context.Context, jobId string, name string) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}

func (s *memoryServer) GetResultImage(ctx context.Context, jobId string, name string) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}

func (s *memoryServer) GetProgressImage(ctx context.Context, jobId string, name string, index int) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}
