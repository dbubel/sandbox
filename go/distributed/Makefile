.PHONY: all proto clean master slave1 slave2

all: proto master slave1 slave2

# Generate protobuf code
proto:
	mkdir -p proto/generated
	protoc --go_out=proto/generated --go-grpc_out=proto/generated --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/distributed.proto

# Run master server
master:
	go run master/master.go

# Run slave server 1
slave1:
	go run slave/slave.go --id=1 --port=5001

# Run slave2 server
slave2:
	go run slave/slave.go --id=2 --port=5002

# Install dependencies
deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Clean generated files
clean:
	rm -rf proto/generated

# Local setup 
setup: deps proto 