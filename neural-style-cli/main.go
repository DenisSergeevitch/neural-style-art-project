package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	"golang.org/x/net/context"

	"github.com/mgilbir/neural-style-art-project/pb"
	"google.golang.org/grpc"
)

var (
	styleFile   = flag.String("style_image", "", "style image")
	contentFile = flag.String("content_image", "", "content image")
	name        = flag.String("name", "", "name")
	grpcConnStr = flag.String("grpc", ":8081", "The gRPC connection string")
)

func main() {
	flag.Parse()

	//TODO: Fix the insecure thingie
	conn, err := grpc.Dial(*grpcConnStr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	style, err := os.Open(*styleFile)
	if err != nil {
		log.Fatal(err)
	}
	defer style.Close()

	styleImg, err := ioutil.ReadAll(style)

	_, styleName := path.Split(*styleFile)
	ext := len(path.Ext(styleName))
	styleName = styleName[:len(styleName)-ext]

	content, err := os.Open(*contentFile)
	if err != nil {
		log.Fatal(err)
	}
	defer content.Close()

	contentImg, err := ioutil.ReadAll(content)
	if err != nil {
		log.Fatal(err)
	}

	cl := pb.NewNeuralStyleImagerClient(conn)

	job := pb.CreateFullJobRequest{
		Name: *name,
		Style: &pb.InputImage{
			Title:  styleName,
			Format: pb.ImageFormat_JPG,
			Image:  styleImg,
		},
		Content: &pb.InputImage{
			Title:  *name,
			Format: pb.ImageFormat_JPG,
			Image:  contentImg,
		},
	}

	ctx := context.Background()
	_, err = cl.CreateFullJob(ctx, &job)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Job submitted")
}
