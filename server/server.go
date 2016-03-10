package server

import (
	"time"

	"github.com/mgilbir/neural-style-art-project/pb"
	_ "github.com/mgilbir/neural-style-art-project/server/statik"
	"golang.org/x/net/context"
)

type Job struct {
	Name           string
	StyleName      string
	StyleImage     []byte
	ContentImage   []byte
	PartialResults [][]byte
	Result         []byte
	LastUpdated    time.Time
}

type jobKey struct {
	ID        string
	Name      string
	Completed bool
}

type Server interface {
	StyleLoader
	ImagerServer
	JobServer
	UIServer
	Closer
}

type StyleLoader interface {
	LoadStyle(filename string) error
}

type ImagerServer interface {
	CreateJob(ctx context.Context, in *pb.CreateJobRequest) (*pb.CreateJobResponse, error)
	CreateFullJob(ctx context.Context, in *pb.CreateFullJobRequest) (*pb.CreateFullJobResponse, error)
}

type JobServer interface {
	RequestJob(ctx context.Context, in *pb.JobRequest) (*pb.Job, error)
	AcknowledgeJob(ctx context.Context, in *pb.JobAck) (*pb.JobAck, error)
	ProgressReport(ctx context.Context, in *pb.JobResult) (*pb.JobProgressResponse, error)
	CompleteJob(ctx context.Context, in *pb.JobResult) (*pb.JobResultResponse, error)
	FailJob(ctx context.Context, in *pb.JobFail) (*pb.JobFail, error)
}

type JobResponse struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Status            string   `json:"status"`
	StyleImageUrl     string   `json:"styleUrl"`
	ContentImageUrl   string   `json:"contentUrl"`
	ProgressImageUrls []string `json:"progressUrls,omitempty"`
	ResultImageUrl    string   `json:"resultUrl,omitempty"`
}

type JobStats struct {
	PendingJobsCount    int `json:"pending"`
	InProgressJobsCount int `json:"inprogress"`
	CompletedJobsCount  int `json:"completed"`
}

type AllJobsResponse struct {
	Jobs  []JobResponse `json:"jobs"`
	Stats JobStats      `json:"stats"`
}

type UIServer interface {
	GetAllJobs(ctx context.Context) (*AllJobsResponse, error)
	GetStyleImage(ctx context.Context, jobId string, name string) ([]byte, error)
	GetContentImage(ctx context.Context, jobId string, name string) ([]byte, error)
	GetResultImage(ctx context.Context, jobId string, name string) ([]byte, error)
	GetProgressImage(ctx context.Context, jobId string, name string, index int) ([]byte, error)
}

type Closer interface {
	Close() error
}
