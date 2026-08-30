# Orbit: gRPC Todo API & Client

Orbit is a complete, lightweight task management backend and CLI tool written in Go. It demonstrates a clean implementation of a gRPC service backed by a zero-configuration SQLite database. With built-in support for cross-compilation, dynamic query filtering, and terminal-friendly ANSI output, it serves as a robust foundation for building cross-platform desktop applications (like Fyne) or SaaS products.

## ✨ Features

* **gRPC Architecture:** Strongly typed, high-performance client-server communication using Protocol Buffers.
* **SQLite Persistence:** Self-contained database (`go-sqlite3`) with dynamic SQL query building for filtering and sorting.
* **Advanced CLI Client:** Command-line interface with subcommands, pagination, and ANSI-colorized output grouping Work and Personal tasks.
* **Cross-Platform Ready:** Includes a `Makefile` configured for native and cross-compiling (Windows/macOS) directly from a Linux host.

## 📂 Project Structure

```text
.
├── client/
│   └── main.go         # Interactive gRPC CLI client
├── proto/
│   ├── todo.proto      # Protocol Buffers definitions
│   ├── todo.pb.go      # Generated message structures
│   └── todo_grpc.pb.go # Generated gRPC client/server code
├── server/
│   └── main.go         # gRPC server and SQLite database logic
├── Makefile            # Build automation and cross-compilation
├── .gitignore          # Version control exclusions
├── go.mod
└── go.sum

⚙️ Prerequisites

    Go (1.18+)

    Protocol Buffers Compiler (protoc)

    protoc-gen-go and protoc-gen-go-grpc plugins

    A C compiler (e.g., gcc) for SQLite CGO bindings.

    For cross-compiling from Linux: mingw-w64 (Windows) and osxcross (macOS).

🚀 Getting Started

1. Clone and Install Dependencies
git clone [https://github.com/yourusername/orbit-todo-grpc.git](https://github.com/yourusername/orbit-todo-grpc.git)
cd orbit-todo-grpc
go mod tidy

2. Generate gRPC Code
If you modify proto/todo.proto, regenerate the Go files:
make proto

3. Run the Server
Starts the gRPC server on :50051 and initializes todos.db:
make run-server

💻 CLI Usage
Open a new terminal window to interact with the API using the client.
Create a Task:
go run client/main.go create -desc "Review architecture" -category "WORK" -due "2026-09-05T12:00:00Z"

List and Filter Tasks:
The list command automatically groups tasks by category and highlights overdue items in red.
# View all with pagination
go run client/main.go list -limit 10 -offset 0

# Filter by search term and sort by due date
go run client/main.go list -search "architecture" -sort "due" -asc

# Show only uncompleted personal tasks
go run client/main.go list -category "PERSONAL" -completed "false"

Update a Task:
go run client/main.go update -id "<UUID>" -completed=true

Delete a Task:
go run client/main.go delete -id "<UUID>"

🛠️ Building Binaries

Use the provided Makefile to compile standalone binaries for the server and client. Because SQLite requires CGO, cross-compiling relies on specific C toolchains.
# Build for your host OS (outputs to bin/)
make build

# Cross-compile for Windows (from Linux)
make build-windows

# Cross-compile for macOS (from Linux)
make build-macos

# Clean up binaries and the local database
make clean