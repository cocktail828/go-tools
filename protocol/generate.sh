#!/bin/bash

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

export PATH="$PATH:$(go env GOPATH)/bin"

mkdir -p ./messagepb
protoc --go_out=./messagepb --go_opt=paths=source_relative \
  --go-grpc_out=./messagepb --go-grpc_opt=paths=source_relative \
  message.proto
