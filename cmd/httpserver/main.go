package main

import (
	"crypto/sha256"
	"fmt"
	"http/internal/headers"
	"http/internal/request"
	"http/internal/response"
	"http/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port uint16 = 42069

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK
		switch {
		case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"):
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org" + target[len("/httpbin"):])
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Delete("Content-length")
				h.Set("Transfer-encoding", "chunked")
				h.Replace("Content-type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(*h)

				fullBody := []byte{}
				for {
					data := make([]byte, 1024)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}
					if n > 0 {
						fullBody = append(fullBody, data[:n]...)
						w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
						w.WriteBody(data[:n])
						w.WriteBody([]byte("\r\n"))
					}
				}
				w.WriteBody([]byte("0\r\n"))
				trailer := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailer.Set("X-Content-SHA256", toStr(out[:]))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailer)
				return
			}

		case req.RequestLine.RequestTarget == "/yourproblem":
			body = respond400()
			status = response.StatusBadRequest
		case req.RequestLine.RequestTarget == "/myproblem":
			body = respond500()
			status = response.StatusInternalServerError
		}
		h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v ", err)
	}
	defer server.Close()
	log.Printf("Server started on port: %v", port)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
