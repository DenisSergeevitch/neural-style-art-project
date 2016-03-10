all:
	find ../ -name "*.pb.go" -exec rm {} \;
	protoc -I . --go_out=plugins=grpc,Mimager.proto=github.com/mgilbir/neural-style-art-project/pb,Mimage.proto=github.com/mgilbir/neural-style-art-project/pb,Mjobs.proto=github.com/mgilbir/neural-style-art-project/pb,import_path=pb:../pb *.proto
