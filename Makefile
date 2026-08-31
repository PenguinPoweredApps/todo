# Variables
PROTO_DIR := proto
SERVER_DIR := server
CLIENT_DIR := client
BIN_DIR := bin

# Default target
.PHONY: all
all: gen build

# Install required gRPC Go plugins
.PHONY: install-tools
install-tools:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate gRPC Go code from .proto files
.PHONY: gen
gen:
	@echo "Generating gRPC Go code..."
	protoc --go_out=. --go_opt=module=github.com/penguinpoweredapps/todo \
	       --go-grpc_out=. --go-grpc_opt=module=github.com/penguinpoweredapps/todo \
	       proto/todo.proto

# Build the client and server binaries
.PHONY: build
build:
	@echo "Building binaries into $(BIN_DIR)/..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./$(SERVER_DIR)
	go build -o $(BIN_DIR)/client ./$(CLIENT_DIR)

# Run the server
.PHONY: run-server
run-server:
	@echo "Starting gRPC server..."
	go run ./$(SERVER_DIR)

# Run the client
.PHONY: run-client
run-client:
	@echo "Starting gRPC client..."
	go run ./$(CLIENT_DIR)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)