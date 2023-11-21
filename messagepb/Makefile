export GO111MODULE=on
export GOPROXY=https://goproxy.cn

gopath=$(shell go env GOPATH)
default:
	rm -rf *.go && protoc -I. --go-grpc_out=:. --go_out=:. message.proto

env:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
