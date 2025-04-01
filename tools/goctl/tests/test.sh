#!/usr/bin/env bash

rm -rf demo && go build -o x ../goctl.go && ./x go --api svc.api --dir demo
# rm -rf demo && go build -o x ../goctl.go && ./x rpc protoc abc_xxx_asd.proto --go_out=./demo/pb --go-grpc_out=./demo/pb --zrpc_out=./demo
