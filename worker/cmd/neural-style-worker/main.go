package main

import (
	"flag"
	"log"

	"golang.org/x/net/context"

	"github.com/mgilbir/neural-style-art-project/worker"
	"google.golang.org/grpc"
)

var (
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

	w := worker.New(conn)

	ctx := context.Background()
	w.Run(ctx)
}
