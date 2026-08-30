# Output binary names
SERVER_OUT = bin/server
CLIENT_OUT = bin/client

.PHONY: all proto build build-linux build-windows build-macos clean

all: proto build

# Generate gRPC code
proto:
	protoc --go_out=. --go-grpc_out=. proto/todo.proto

# Default build (Host OS)
build:
	go build -o $(SERVER_OUT) ./server
	go build -o $(CLIENT_OUT) ./client

# Native Linux Build
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(SERVER_OUT)-linux ./server
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(CLIENT_OUT)-linux ./client

# Cross-compile for Windows from Linux (Requires mingw-w64)
# Ubuntu/Debian setup: sudo apt install gcc-mingw-w64
build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(SERVER_OUT).exe ./server
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(CLIENT_OUT).exe ./client

# Cross-compile for macOS from Linux (Requires osxcross toolchain)
build-macos:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC=o64-clang go build -o $(SERVER_OUT)-darwin ./server
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC=o64-clang go build -o $(CLIENT_OUT)-darwin ./client

# Run the server locally
run-server:
	go run server/main.go

# Clean up binaries and local database
clean:
	rm -rf bin/
	rm -f todos.db