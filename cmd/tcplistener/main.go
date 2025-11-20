package main

import (
	"fmt"
	"http/internal/request"
	"log"
	"net"
	"os"
)

func getReadFromFile() *os.File {
	fmt.Println("I hope I get the job")
	f, err := os.Open("message.txt")
	if err != nil {
		log.Fatal("error", err)
	}
	return f
}

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	out := make(chan string, 1)
// 	go func() {
// 		defer f.Close()
// 		defer fmt.Println("Channel closed") // This will be printed after the channel is closed.
// 		defer close(out)
// 		str := ""
// 		for {
// 			data := make([]byte, 8)
// 			n, err := f.Read(data)
// 			if err != nil {
// 				break
// 			}
// 			data = data[:n]
// 			for len(data) > 0 {
// 				if i := bytes.IndexByte(data, '\n'); i != -1 {
// 					str += string(data[:i])
// 					out <- str
// 					str = ""
// 					data = data[i+1:]
// 				} else {
// 					str += string(data)
// 					data = nil
// 				}
// 			}
// 		}

// 		if len(str) != 0 {
// 			out <- str
// 		}
// 	}()
// 	return out
// }

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error: ", err)
	}

	for {
		conn, err := listener.Accept()
		if conn != nil {
			fmt.Println("Connection Accepted")
		}
		if err != nil {
			log.Fatal("error: ", err)
		}
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error: ", err)
		}
		fmt.Printf("Request line: \n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf(" - Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf(" - Version: %s\n", r.RequestLine.HttpVersion)
	}

	// *** For Reading from file ***
	// lines := getLinesChannel(getReadFromFile())
	// for line := range lines {
	// 	fmt.Printf("%s\n", line)
	// }

}
