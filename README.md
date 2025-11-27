# HTTP From Scratch

A from-scratch implementation of TCP/HTTP protocols in Go, building a complete HTTP/1.1 server without using Go's standard `net/http` package for core HTTP parsing and handling.

## 🎯 Project Goals

- **Understand HTTP at the protocol level**: Parse raw TCP streams into HTTP requests
- **Implement streaming parsers**: Handle data as it arrives, not all at once
- **Learn state machines**: Use state-driven parsing for robustness
- **Master chunked encoding**: Stream responses without knowing size upfront
- **Explore concurrency**: Handle multiple connections simultaneously

## 🏗️ Architecture

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

## 📁 Project Structure

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

## 🔑 Key Components

### 1. **Headers Package** (`internal/headers/`)

Parses and manages HTTP headers with RFC compliance.

**Features**:
- Case-insensitive storage (normalized to lowercase)
- Multi-value header support (comma-separated)
- Token validation for header names
- Streaming parser (handles partial data)

```go
headers := headers.NewHeaders()
headers.Set("Content-Type", "application/json")
headers.Set("Content-Type", "charset=utf-8")  // Appends
value, ok := headers.Get("content-type")      // Case-insensitive
// value = "application/json,charset=utf-8"
```

### 2. **Request Package** (`internal/request/`)

State machine-based HTTP request parser.

**States**:
- `StateInit`: Parse request line (`GET /path HTTP/1.1`)
- `StateHeaders`: Parse headers until empty line
- `StateBody`: Read body based on `Content-Length`
- `StateDone`: Parsing complete

**Why State Machine?**: Handles streaming data gracefully. If data is incomplete, state persists until more arrives.

```go
request, err := request.RequestFromReader(conn)
fmt.Println(request.RequestLine.Method)        // "GET"
fmt.Println(request.RequestLine.RequestTarget) // "/path"
fmt.Println(request.Body())                    // Request body
```

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

## 🚀 HTTP Server Features

The main HTTP server (`cmd/httpserver/`) demonstrates advanced HTTP features:

### Routes

| Route | Description | Features |
|-------|-------------|----------|
| `/` | Success page | Returns 200 with HTML |
| `/httpbin/*` | Proxy to httpbin.org | Chunked transfer encoding, trailers, SHA256 checksum |
| `/video` | Serve MP4 file | Binary data streaming, `Content-Type: video/mp4` |
| `/yourproblem` | Client error demo | Returns 400 Bad Request |
| `/myproblem` | Server error demo | Returns 500 Internal Server Error |

### Advanced Features

#### **Chunked Transfer Encoding** (`/httpbin/*`)

Streams responses without knowing total size upfront:

```
HTTP/1.1 200 OK
Transfer-Encoding: chunked
Trailer: X-Content-SHA256

5\r\n          ← Chunk size (hex)
Hello\r\n      ← Chunk data
6\r\n
World!\r\n
0\r\n          ← End marker
X-Content-SHA256: abc123...\r\n  ← Trailer
\r\n
```

**Implementation**:
```go
for {
    data := make([]byte, 1024)
    n, _ := res.Body.Read(data)
    
    w.WriteBody(fmt.Sprintf("%x\r\n", n))  // Size in hex
    w.WriteBody(data[:n])                   // Data
    w.WriteBody([]byte("\r\n"))            // Delimiter
}
w.WriteBody([]byte("0\r\n"))  // End
```

#### **HTTP Trailers**

Headers sent after the body (useful for checksums computed during streaming):

```go
h.Set("Trailer", "X-Content-SHA256")  // Announce trailer
// ... send chunked body ...
trailer := headers.NewHeaders()
trailer.Set("X-Content-SHA256", sha256sum)
w.WriteHeaders(*trailer)
```

#### **Binary Data Streaming** (`/video`)

Serves video files with proper content type:

```go
f, _ := os.ReadFile("assets/vim.mp4")
h.Replace("Content-Type", "video/mp4")
h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))
w.WriteBody(f)
```

## 🧪 Testing

Comprehensive test suite with network simulation:

```bash
go test ./internal/headers -v
go test ./internal/request -v
```

**chunkReader**: Simulates network by reading N bytes at a time:
```go
reader := &chunkReader{
    data:            "GET / HTTP/1.1\r\n...",
    numBytesPerRead: 3,  // Simulate 3-byte chunks
}
request, _ := RequestFromReader(reader)
```

## 🏃 Running

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

## 📚 Key Concepts

### **Streaming vs Buffering**

**Buffering** (memory intensive):
```go
data, _ := ioutil.ReadAll(reader)  // Load entire file
parse(data)
```

**Streaming** (this project):
```go
buf := make([]byte, 8192)
for {
    n, _ := reader.Read(buf)  // Read chunk
    process(buf[:n])          // Process immediately
}
```

### **State Machine Pattern**

Handles partial data elegantly:
```go
for !done {
    switch state {
    case StateInit:
        // Try to parse request line
        // If incomplete, wait for more data
    case StateHeaders:
        // Parse headers incrementally
    }
}
```

### **HTTP Protocol Structure**

```
GET /path HTTP/1.1\r\n          ← Request line
Host: example.com\r\n           ← Headers
Content-Length: 13\r\n
\r\n                            ← Empty line
Hello, World!                   ← Body
```

**`\r\n`**: CRLF (Carriage Return + Line Feed) - HTTP line delimiter

## 🛠️ Dependencies

- `github.com/stretchr/testify` - Testing assertions

## 📖 Learning Path

1. **Start with UDP sender**: Understand basic networking
2. **Explore TCP listener**: See raw HTTP requests
3. **Study headers package**: Learn HTTP header parsing
4. **Dive into request parser**: Understand state machines
5. **Build with HTTP server**: See it all come together

## 🎓 What You'll Learn

- TCP socket programming in Go
- HTTP/1.1 protocol specification
- Streaming data processing
- State machine design patterns
- Concurrent connection handling
- Chunked transfer encoding
- HTTP trailers and checksums
- Binary data handling

## 🤝 Contributing

This is a learning project! Feel free to:
- Add new routes
- Implement HTTP/2 features
- Add middleware support
- Improve error handling
- Optimize performance

## 📝 License

Open source - use for learning and experimentation!

---

**Built with ❤️ to understand HTTP from the ground up**
