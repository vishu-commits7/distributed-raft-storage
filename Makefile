.PHONY: help protoc build test run clean

help:
	@echo "Available targets:"
	@echo "  make protoc       - Generate protobuf Go code"
	@echo "  make build        - Build server and client binaries"
	@echo "  make test         - Run tests"
	@echo "  make run          - Run a single node"
	@echo "  make clean        - Clean build artifacts"

protoc:
	@command -v protoc >/dev/null 2>&1 || { echo "protoc is not installed"; exit 1; }
	@command -v protoc-gen-go >/dev/null 2>&1 || go install github.com/golang/protobuf/protoc-gen-go@latest
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc --go_out=. --go-grpc_out=. proto/storage.proto

build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

test:
	go test -v -race ./...

run:
	go run cmd/server/main.go -id node1 -bootstrap

clean:
	rm -rf bin/
	rm -rf /tmp/raft-node*
	go clean
