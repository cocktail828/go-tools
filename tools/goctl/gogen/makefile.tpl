
SHELL = /bin/bash
export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on
export GOSUMDB=off

# change me to bump version
APPVER ?= v0.0.1

.PHONY: init ci lint vendor run clean

init:
	go mod init {{.mod}}
	go mod tidy

ci: clean generate caesar
	go build -tags=jsoniter -ldflags " \
		-X '{{.mod}}/vars.GitTag=$(shell git tag --sort=version:refname | tail -n 1)' \
		-X '{{.mod}}/vars.CommitLog=$(shell git log --pretty=oneline -n 1)' \
		-X '{{.mod}}/vars.BuildTime=$(shell date +'%Y.%m.%d.%H%M%S')' \
		-X '{{.mod}}/vars.GoVersion=$(shell go version)' \
		-X '{{.mod}}/vars.AppVersion=$(APPVER)' \
	" -o objs/$@ main.go

lint:
	go fmt ./...
	go generate

vendor:
	go mod tidy && go mod vendor

run: ci
	# change me on need
	pushd objs && ./{{.mod}} && popd

clean:
	rm -rf objs/*
