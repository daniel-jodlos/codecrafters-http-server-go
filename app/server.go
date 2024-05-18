package main

import (
	"fmt"
	"log"

	// Uncomment this block to pass the first stage
	 "net"
	 "os"
)

const CRLF string = "\r\n"

func buildHttpResponse(status int, reason string, headers []any) string {
	if len(headers) > 0 {
		log.Panic("Headers are not yet supported")
	}

	return fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", status, reason)
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	 l, err := net.Listen("tcp", "0.0.0.0:4221")
	 if err != nil {
	 	fmt.Println("Failed to bind to port 4221")
	 	os.Exit(1)
	 }

	 conn, err := l.Accept()
	 if err != nil {
	 	fmt.Println("Error accepting connection: ", err.Error())
	 	os.Exit(1)
	 }

	_, err = conn.Write([]byte(buildHttpResponse(200, "OK", make([]any, 0))))
	if err != nil {
		fmt.Println("failed to send data")
		os.Exit(1)
	}
}
