package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	port := ":42069"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	fmt.Println("Listening for TCP traffic on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		// outputs := getLinesChannel(conn)
		// for line := range outputs {
		// 	fmt.Println(line)
		// }
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\nHeaders:\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion)
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}
		fmt.Printf("Body:\n%s\n", string(req.Body))
		// when outputs channel is closed, the connection is closed
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	outputCh := make(chan string)
// 	go func() {
// 		defer f.Close()
// 		defer close(outputCh)
// 		output := ""
// 		for {
// 			bytes := make([]byte, 8)
// 			n, err := f.Read(bytes)
// 			if err != nil {
// 				if errors.Is(err, io.EOF) {
// 					if len(output) > 0 {
// 						outputCh <- output
// 					}
// 					break
// 				}
// 				fmt.Printf("error: %v\n", err.Error())
// 				break
// 			}
// 			output += string(bytes[:n])
// 			parts := strings.Split(output, "\n")
// 			for i := 0; i < len(parts)-1; i++ {
// 				outputCh <- parts[i]
// 			}
// 			output = parts[len(parts)-1]
// 		}
// 	}()
// 	return outputCh
// }
