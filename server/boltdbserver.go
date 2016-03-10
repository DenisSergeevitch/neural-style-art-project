package server

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/mgilbir/neural-style-art-project/pb"
	_ "github.com/mgilbir/neural-style-art-project/server/statik"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

type boltDbServer struct {
	db *bolt.DB
}

func NewBoltDbServer(filename string) (Server, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = InitializeBoltDb(db)
	if err != nil {
		return nil, err
	}

	return &boltDbServer{
		db: db,
	}, nil
}

func (s *boltDbServer) Close() error {
	return s.db.Close()
}

func (s *boltDbServer) LoadStyle(filename string) error {
	return fmt.Errorf("Not implemented")
}

func (s *boltDbServer) CreateFullJob(ctx context.Context, in *pb.CreateFullJobRequest) (*pb.CreateFullJobResponse, error) {
	return &pb.CreateFullJobResponse{}, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) CreateJob(ctx context.Context, in *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	var keyBuffer bytes.Buffer
	enc := gob.NewEncoder(&keyBuffer)
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	err = enc.Encode(jobKey{
		ID:        id.String(),
		Name:      in.Name,
		Completed: false,
	})

	if err != nil {
		return nil, err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(NewBucket))
		err := b.Put(keyBuffer.Bytes(), []byte("42"))
		return err
	})

	return &pb.CreateJobResponse{}, err
}

func (s *boltDbServer) RequestJob(ctx context.Context, in *pb.JobRequest) (*pb.Job, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) AcknowledgeJob(ctx context.Context, in *pb.JobAck) (*pb.JobAck, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) ProgressReport(ctx context.Context, in *pb.JobResult) (*pb.JobProgressResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) CompleteJob(ctx context.Context, in *pb.JobResult) (*pb.JobResultResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) FailJob(ctx context.Context, in *pb.JobFail) (*pb.JobFail, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) GetAllJobs(ctx context.Context) (*AllJobsResponse, error) {
	return &AllJobsResponse{}, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) GetStyleImage(ctx context.Context, jobId string, name string) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) GetContentImage(ctx context.Context, jobId string, name string) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) GetResultImage(ctx context.Context, jobId string, name string) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}

func (s *boltDbServer) GetProgressImage(ctx context.Context, jobId string, name string, index int) ([]byte, error) {
	return []byte{}, fmt.Errorf("Not implemented")
}
