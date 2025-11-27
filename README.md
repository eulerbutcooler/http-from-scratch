# HTTP From Scratch

A from-scratch implementation of TCP/HTTP protocols in Go, building a complete HTTP/1.1 server without using Go's standard `net/http` package for core HTTP parsing and handling.

## Project Goals

- **Understand HTTP at the protocol level**: Parse raw TCP streams into HTTP requests
- **Implement streaming parsers**: Handle data as it arrives, not all at once
- **Learn state machines**: Use state-driven parsing for robustness
- **Master chunked encoding**: Stream responses without knowing size upfront
- **Explore concurrency**: Handle multiple connections simultaneously

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Server                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │  TCP Listener (port 42069)                       │  │
│  │  ├─ Accept connections                           │  │
│  │  └─ Spawn goroutine per connection               │  │
│  └──────────────────────────────────────────────────┘  │
│                         ↓                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Request Parser (State Machine)                  │  │
│  │  ├─ StateInit:    Parse request line            │  │
│  │  ├─ StateHeaders: Parse headers                 │  │
│  │  ├─ StateBody:    Read body (Content-Length)    │  │
│  │  └─ StateDone:    Complete                      │  │
│  └──────────────────────────────────────────────────┘  │
│                         ↓                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Handler Function                                │  │
│  │  ├─ Route matching                               │  │
│  │  ├─ Business logic                               │  │
│  │  └─ Response generation                          │  │
│  └──────────────────────────────────────────────────┘  │
│                         ↓                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Response Writer                                 │  │
│  │  ├─ Write status line                           │  │
│  │  ├─ Write headers                                │  │
│  │  └─ Write body (standard or chunked)            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Project Structure

```
http-from-scratch/
├── cmd/
│   ├── httpserver/     # Full HTTP/1.1 server with routing
│   ├── tcplistener/    # Basic TCP listener (learning tool)
│   └── udpsender/      # UDP sender example
├── internal/
│   ├── headers/        # HTTP header parsing & management
│   ├── request/        # HTTP request parsing (state machine)
│   ├── response/       # HTTP response writing
│   └── server/         # TCP server & connection handling
├── assets/             # Static files (test video)
└── message.txt         # Test data
```

## Key Components

### 1. **Headers Package** (`internal/headers/`)

Parses and manages HTTP headers with RFC compliance.

**Features**:
- Case-insensitive storage (normalized to lowercase)
- Multi-value header support (comma-separated)
- Token validation for header names
- Streaming parser (handles partial data)

### 2. **Request Package** (`internal/request/`)

State machine-based HTTP request parser.

**States**:
- `StateInit`: Parse request line (`GET /path HTTP/1.1`)
- `StateHeaders`: Parse headers until empty line
- `StateBody`: Read body based on `Content-Length`
- `StateDone`: Parsing complete

### 3. **Response Package** (`internal/response/`)

Writes HTTP responses to TCP connections.

```go
writer := response.NewWriter(conn)
writer.WriteStatusLine(response.StatusOK)
writer.WriteHeaders(headers)
writer.WriteBody([]byte("Hello, World!"))
```

### 4. **Server Package** (`internal/server/`)

TCP server with concurrent connection handling.

```go
server, _ := server.Serve(42069, func(w *response.Writer, req *request.Request) {
    // Handle request
    w.WriteStatusLine(response.StatusOK)
    w.WriteHeaders(*response.GetDefaultHeaders(0))
    w.WriteBody([]byte("Response"))
})
defer server.Close()
```

## HTTP Server Features

The main HTTP server (`cmd/httpserver/`) has these features:

### Routes

| Route | Description | Features |
|-------|-------------|----------|
| `/` | Success page | Returns 200 with HTML |
| `/httpbin/*` | Proxy to httpbin.org | Chunked transfer encoding, trailers, SHA256 checksum |
| `/video` | Serve MP4 file | Binary data streaming, `Content-Type: video/mp4` |
| `/yourproblem` | Client error demo | Returns 400 Bad Request |
| `/myproblem` | Server error demo | Returns 500 Internal Server Error |

### Other features

#### **Chunked Transfer Encoding** (`/httpbin/*`)

Streams responses without knowing total size upfront:

#### **HTTP Trailers**

Headers sent after the body (useful for checksums computed during streaming):

#### **Binary Data Streaming** (`/video`)

Serves video files with proper content type:

## Testing

Comprehensive test suite with network simulation:

```bash
go test ./internal/headers -v
go test ./internal/request -v
```
or simply do
```bash
go test ./...
```

## Running

### HTTP Server

```bash
# Start server
go run cmd/httpserver/main.go

# Test routes
curl http://localhost:42069/
curl http://localhost:42069/httpbin/get
curl http://localhost:42069/video --output video.mp4
curl http://localhost:42069/yourproblem
```

### TCP Listener (Debug Tool)

```bash
# Terminal 1: Start listener
go run cmd/tcplistener/main.go

# Terminal 2: Send request
curl http://localhost:42069/test
```

### UDP Sender

```bash
# Terminal 1: Listen for UDP
nc -u -l 42068

# Terminal 2: Send messages
go run cmd/udpsender/main.go
```

## Dependencies

- `github.com/stretchr/testify` - Testing assertions
---

