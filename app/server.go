package main

import (
	"fmt"
	"strings"

	"flag"
	// Uncomment this block to pass the first stage
	"bytes"
	"compress/gzip"
	"net"
	"os"
)

var dirFlag = flag.String("directory", "example", "provide the directory path")

const CRLF string = "\r\n"

type Headers map[string]string

func (h *Headers) toString() string {
	result := ""

	for k, v := range *h {
		result += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	return result
}

func (h *Headers) get(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := (*h)[key]
	return value, ok
}

type RequestHeaders Headers

func (h *Headers) getAcceptedEncoding() (string, bool) {
	fmt.Println(*h)
	header, ok := h.get("Accept-Encoding")

	if !ok {
		return "", false
	}

	encodings := strings.Split(header, ",")

	for _, encoding := range encodings {
		encoding = strings.Trim(encoding, " ")
		if isSupportedEncoding(encoding) {
			return encoding, true
		}
	}

	return "", false
}

type HttpRequest struct {
	method      string
	url         string
	body        string
	httpVersion string
	headers     Headers
}

func HttpRequestFromBytes(bytes []byte) HttpRequest {
	x := string(bytes)
	parts := strings.Split(x, CRLF)
	headerParts := strings.Split(parts[0], " ")

	headers := make(Headers)

	for _, line := range parts[1:] {
		fmt.Println(line)
		line = strings.TrimSuffix(line, CRLF)

		if line == "" {
			break
		}

		parts := strings.SplitN(line, " ", 2)
		key := strings.ToLower(parts[0])
		key = strings.TrimSuffix(key, ":")
		headers[key] = strings.Trim(parts[1], " ")
	}

	return HttpRequest{
		method:      headerParts[0],
		url:         headerParts[1],
		body:        parts[2+len(headers)],
		httpVersion: strings.TrimPrefix(headerParts[2], "HTTP/"),
		headers:     headers,
	}
}

func reasonForCode(code uint) string {
	switch code {
	case 200:
		return "OK"
	case 404:
		return "Not Found"
	case 201:
		return "Created"
	default:
		return "OK"
	}
}

type HttpResponse struct {
	body    []byte
	headers Headers
	status  uint
}

func NewHttpResponse() HttpResponse {
	return HttpResponse{status: 404, body: make([]byte, 0), headers: make(Headers)}
}

func NewHttpResponseWithBody(body string) HttpResponse {
	response := NewHttpResponse()
	response.status = 200
	response.headers["Content-Type"] = "text/plain"
	response.headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	response.body = []byte(body)
	return response
}

func NewHttpResponseWithFile(file *os.File) HttpResponse {
	response := NewHttpResponse()
	response.status = 200
	response.headers["Content-Type"] = "application/octet-stream"
	fInfo, err := file.Stat()

	if err != nil {
		fmt.Println("failed to stat the file")
		os.Exit(0)
	}

	response.body = make([]byte, fInfo.Size())
	read, err := file.Read(response.body)
	response.headers["Content-Length"] = fmt.Sprintf("%d", read)

	if err != nil {
		fmt.Println("Failed to load file")
		os.Exit(1)
	}

	return response
}

func (h *HttpResponse) toString() string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", h.status, reasonForCode(h.status), h.headers.toString(), h.body)

}

func (h *HttpResponse) setContentEncoding(encoding string) {
	h.headers["Content-Encoding"] = encoding

	if encoding == "gzip" {
		buf := &bytes.Buffer{}
		writer := gzip.NewWriter(buf)
		writer.Write(h.body)
		writer.Close()
		h.body = buf.Bytes()
		h.headers["Content-Length"] = fmt.Sprintf("%d", len(h.body))
	} else {
		panic("Unsupported encoding")
	}
}

func isSupportedEncoding(encoding string) bool {
	return encoding == "gzip"
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	readBytes, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Failed to read bytes")
		os.Exit(1)
	}

	request := HttpRequestFromBytes(buffer[:readBytes])

	var response = NewHttpResponse()

	switch {
	case request.url == "/":
		response.status = 200
	case request.url == "/user-agent":
		body, _ := request.headers.get("User-Agent")
		response = NewHttpResponseWithBody(body)
	case strings.HasPrefix(request.url, "/echo/"):
		body := strings.TrimPrefix(request.url, "/echo/")
		response = NewHttpResponseWithBody(body)
	case request.method == "GET" && strings.HasPrefix(request.url, "/files/"):
		file, err := os.Open(*dirFlag + "/" + strings.TrimPrefix(request.url, "/files/"))

		if err != nil {
			response = NewHttpResponse()
		} else {
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					fmt.Println("Failed to close the file")
				}
			}(file)
			response = NewHttpResponseWithFile(file)
		}
	case request.method == "POST" && strings.HasPrefix(request.url, "/files/"):
		response.status = 201
		file, err := os.Create(*dirFlag + "/" + strings.TrimPrefix(request.url, "/files/"))

		if err != nil {
			fmt.Println("Failed to open file")
			response.status = 500
		} else {
			_, err = file.Write([]byte(request.body))

			if err != nil {
				fmt.Println("Failed to save to the file")
				response.status = 500
			}
		}
	}

	if encoding, ok := request.headers.getAcceptedEncoding(); ok {
		response.setContentEncoding(encoding)
	}

	_, err = conn.Write([]byte(response.toString()))
	if err != nil {
		fmt.Println("failed to send data")
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}
