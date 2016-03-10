package worker

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mgilbir/neural-style-art-project/pb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Worker struct {
	conn          *grpc.ClientConn
	client        pb.NeuralStyleWorkerClient
	basedir       string
	maxIterations int32
}

func New(conn *grpc.ClientConn) *Worker {
	return &Worker{
		conn:          conn,
		client:        pb.NewNeuralStyleWorkerClient(conn),
		maxIterations: 500,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	emptyJob := pb.Job{}
	for {
		//TODO: cancel on the context?

		//Get new job
		job, err := w.client.RequestJob(ctx, &pb.JobRequest{})
		if err != nil {
			log.Println(err)
		}

		if *job == emptyJob {
			//Nothing to see here, keep moving but not too quickly...
			log.Println("No jobs availabla, wait and retry")
			time.Sleep(10 * time.Second)
			continue
		}

		jobDir := getDirectory(w.basedir, job.Id, job.Name)

		//TODO: cleanup
		styleFilename, err := prepareFilename(jobDir, job.Style.Title, job.Style.Format)
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}
		err = ioutil.WriteFile(styleFilename, job.Style.Image, 0755)
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}

		contentFilename, err := prepareFilename(jobDir, job.Content.Title, job.Content.Format)
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}
		err = ioutil.WriteFile(contentFilename, job.Content.Image, 0755)
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}

		//Run job
		iterations := int32(0)

		cmd := exec.Command("th", "neural_style.lua",
			"-style_image", styleFilename,
			"-content_image", contentFilename,
			"-backend", "cudnn", "-cudnn_autotune",
			"-gpu", "0",
			"-print_iter", "1",
			"-num_iterations", strconv.Itoa(int(w.maxIterations)),
		)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}
		err = cmd.Start()
		if err != nil {
			log.Println(err)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}

		//Report progress
		stdoutScanner := bufio.NewScanner(stdout)
		go func() {
			for stdoutScanner.Scan() {
				txt := stdoutScanner.Text()
				txt = strings.TrimSpace(txt)
				tokens := strings.SplitN(txt, " ", -1)
				if strings.Contains(tokens[0], "Iteration") {
					i, err := strconv.Atoi(tokens[1])
					if err != nil {
						log.Printf("Problem parsing iteration %q. %v", txt, err)
					}
					iterations = int32(i)

					if i > 50 && i%100 == 0 {
						go func(i int32) {
							outfile := "out.png"
							if i != w.maxIterations {
								ii := i / 100
								outfile = fmt.Sprintf("out_%d00.png", ii)
							}

							//Give time to the OS to save the file
							time.Sleep(3 * time.Second)
							dest := path.Join(jobDir, outfile)
							err := os.Rename(outfile, dest)
							if err != nil {
								log.Printf("Error while moving file %q to %q. %v", outfile, dest, err)
							}

							//Report progress
							f, err := os.Open(dest)
							if err != nil {
								log.Printf("Error opening moved file. %v", err)
							}

							img, err := ioutil.ReadAll(f)
							if err != nil {
								log.Printf("Error reading moved file. %v", err)
							}
							f.Close()

							msg := pb.JobResult{
								Id:            job.Id,
								Name:          job.Name,
								ProgressCount: i,
								Format:        pb.ImageFormat_PNG,
								Image:         img,
							}
							if i != w.maxIterations {
								w.client.ProgressReport(ctx, &msg)
							} else {
								w.client.CompleteJob(ctx, &msg)
							}
						}(int32(i))
					}
				}
			}
		}()

		stderrScanner := bufio.NewScanner(stderr)
		go func() {
			for stderrScanner.Scan() {
				fmt.Printf("ERRORS | %s\n", stderrScanner.Text())
			}
		}()

		err = cmd.Wait()
		if err != nil {
			log.Println(err)
		}

		if iterations != w.maxIterations {
			log.Printf("Failed processing. Expected %d iterations. Saw %d", w.maxIterations, iterations)
			w.client.FailJob(ctx, &pb.JobFail{Id: job.Id, Name: job.Name})
			continue
		}
	}

	return nil
}

func getDirectory(basedir string, id string, name string) string {
	return path.Join(basedir, name, id)
}

func prepareFilename(dir string, filename string, format pb.ImageFormat) (string, error) {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "/tmp/dump", err
	}
	extension := ""
	switch format {
	case pb.ImageFormat_JPG:
		extension = ".jpg"
	case pb.ImageFormat_PNG:
		extension = ".png"
	}
	return path.Join(dir, filename+extension), nil
}
