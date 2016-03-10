package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/mgilbir/neural-style-art-project/pb"
	"github.com/mgilbir/neural-style-art-project/server"
	"google.golang.org/grpc"
)

var (
	outputDir    = flag.String("output", "output", "output directory")
	dbFile       = flag.String("db", "neural-style.boltdb", "The boltdb store where the images are persisted")
	httpConnStr  = flag.String("http", ":9081", "The HTTP connection string")
	grpcConnStr  = flag.String("grpc", ":8081", "The gRPC connection string")
	stylesConfig = flag.String("styles", "styles.json", "The file with the styles")
)

func main() {
	flag.Parse()
	var err error

	// httpfs, err := fs.New()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err = os.MkdirAll(*outputDir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	var s server.Server

	s, err = server.NewMemoryServer(*outputDir)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	f, err := os.Open(*stylesConfig)
	if err != nil {
		log.Fatal(err)
	}

	var styles []string
	d := json.NewDecoder(f)
	err = d.Decode(&styles)
	if err != nil {
		log.Fatal(err)
	}

	f.Close()

	countStyles := 0
	for _, style := range styles {
		if err := s.LoadStyle(style); err != nil {
			log.Fatal(err)
		}
		countStyles++
	}

	log.Printf("Loaded %d styles\n", countStyles)

	errC := make(chan error)

	http.Handle("/", http.FileServer(http.Dir(*outputDir)))
	go func(errC chan error) {
		if err := http.ListenAndServe(*httpConnStr, nil); err != nil {
			errC <- err
		}
	}(errC)

	// imageUpdaters := ImageUpdaters{}
	//
	// http.Handle("/", http.StripPrefix("/files/", http.FileServer(httpfs)))
	// http.HandleFunc("/ws", imageUpdaters.serveWs)
	// go func(errC chan error) {
	// 	if err := http.ListenAndServe(*addr, nil); err != nil {
	// 		errC <- err
	// 	}
	// }(errC)

	lis, err := net.Listen("tcp", *grpcConnStr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gs := grpc.NewServer()
	pb.RegisterNeuralStyleImagerServer(gs, s)
	pb.RegisterNeuralStyleWorkerServer(gs, s)

	go func(errC chan error) {
		errC <- gs.Serve(lis)
	}(errC)

	if err := <-errC; err != nil {
		log.Fatal(err)
	}
}
