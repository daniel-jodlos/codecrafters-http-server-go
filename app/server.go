package main

import (
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	 "net"
	 "os"
)

const CRLF string = "\r\n"

func buildHttpResponse(status uint, reason string, headers Headers, body string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", status, reason, headers.toString(), body)
}

type Headers map[string]string

func (h Headers) toString() string {
	result := ""

	for k, v := range h {
		result += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	return result
}

type HttpRequest struct {
	method string
	url string
	body string
	httpVersion string
	headers Headers
}

func HttpRequestFromBytes(bytes []byte) HttpRequest {
	x := string(bytes)
	parts := strings.Split(x, CRLF)
	headerParts := strings.Split(parts[0], " ")

	headers := make(Headers)

	for _, line := range parts[1:] {
		line = strings.TrimSuffix(line, CRLF)

		if line == "" {
			break
		}

		parts := strings.Split(line, " ")
		headers[parts[0]] = strings.Trim(parts[1], " ")
	}

	return HttpRequest{
		method: headerParts[0],
		url: headerParts[1],
		body:  parts[2 + len(headers)],
		httpVersion: strings.TrimPrefix(headerParts[2], "HTTP/"),
		headers: headers,
	}
}

func reasonForCode(code uint) string {
	switch code {
	case 200:
		return "OK"
	case 404:
		return "Not Found"
	default:
		panic("Unknown code")
	}
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

	 buffer := make([]byte, 1024)
	 readBytes, err := conn.Read(buffer)

	 if err != nil {
		 fmt.Println("Failed to read bytes")
		 os.Exit(1)
	 }

	 request := HttpRequestFromBytes(buffer[:readBytes])
	 fmt.Println(request)

	 var status uint = 200
	 var body string = ""
	 headers := make(Headers)

	switch {
	case request.url == "/":
		status = 200
	case strings.HasPrefix(request.url, "/echo/"):
		body = strings.TrimPrefix(request.url, "/echo/")
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	default:
		status = 404
	}

	_, err = conn.Write([]byte(buildHttpResponse(status, reasonForCode(status), headers, body)))
	if err != nil {
		fmt.Println("failed to send data")
		os.Exit(1)
	}
}
