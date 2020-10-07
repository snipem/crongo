# run: make
SHELL := /bin/bash
.PHONY: test

deps:
	go mod download

build: clean
	go build -o ./TODO TODOPATH

test:
	go test -covermode=count -coverprofile=coverage.tmp ./...

lint:
	golangci-lint

