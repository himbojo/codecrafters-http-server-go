package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
)

var routes = []string{"/", "/home", "/about"}

type ResponseStatusLine struct {
	HTTPVersion          string
	StatusCode           int
	OptionalReasonPhrase string
}

func (statusLine ResponseStatusLine) ToString() string {
	return strings.Join([]string{statusLine.HTTPVersion, strconv.Itoa(statusLine.StatusCode), statusLine.OptionalReasonPhrase}, " ") + "\r\n"
}

type RequestStatusLine struct {
	HTTPMethod    string
	RequestTarget string
	HTTPVersion   string
}

type Request struct {
	StatusLine RequestStatusLine
	Headers    map[string]string
	Body       []byte
}

// Error implements error.
func (r Request) Error() string {
	panic("unimplemented")
}

func handleRequest(conn net.Conn) (Request, error) {
	reader := bufio.NewReader(conn)

	// Read the request line
	requestLineString, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, fmt.Errorf("error reading request line: %w", err)
	}
	// declare and print request line
	requestLineArray := strings.Split(requestLineString, " ")
	requestLine := RequestStatusLine{
		HTTPMethod:    requestLineArray[0],
		RequestTarget: requestLineArray[1],
		HTTPVersion:   requestLineArray[2],
	}

	fmt.Println("\nRequest Line:")
	fmt.Println("HTTP Method:", requestLine.HTTPMethod)
	fmt.Println("Request Target:", requestLine.RequestTarget)
	fmt.Println("HTTP Version:", requestLine.HTTPVersion)

	// Read headers until the last \r\n
	headerMap := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Request{}, fmt.Errorf("error reading headers: %w", err)
		}
		if line == "\r\n" {
			break
		}
		parts := strings.SplitN(line, " ", 2)
		key := parts[0][:len(parts[0])-1]
		value := parts[1]
		headerMap[key] = value
	}

	// print all header key value pairs
	fmt.Println("Header:")
	for key, value := range headerMap {
		fmt.Printf("%s: %s", key, value)
	}

	// Read the rest of the data into the body
	var body []byte
	// buffer := make([]byte, 1024)
	// for {
	// 	n, err := conn.Read(buffer)
	// 	if err != nil {
	// 		fmt.Println("Error reading:", err)
	// 		break
	// 	}
	// 	body = append(body, buffer[:n]...)
	// }

	// print the body if available
	fmt.Println("\nBody:")
	fmt.Print(string(body))

	request := Request{
		StatusLine: requestLine,
		Headers:    headerMap,
		Body:       body,
	}

	return request, nil
}

func handleResponse(conn net.Conn, request Request) error {
	// Status Line
	responseStatusLine := ResponseStatusLine{
		HTTPVersion:          "HTTP/1.1",
		StatusCode:           200,
		OptionalReasonPhrase: "OK",
	}

	// Check if target is in route
	if !slices.Contains(routes, request.StatusLine.RequestTarget) {
		responseStatusLine.StatusCode = 404
		responseStatusLine.OptionalReasonPhrase = "Not Found"
	}
	// Headers
	headers := "\r\n"
	// Response Body
	responseBody := ""
	httpResponse := strings.Join([]string{responseStatusLine.ToString(), headers, responseBody}, "")
	fmt.Printf("\nResponse:\n%q\n", httpResponse)

	httpResponseBytes := []byte(httpResponse)
	// send response
	_, err := conn.Write(httpResponseBytes)
	if err != nil {
		return fmt.Errorf("read timeout: %w", err)
	}
	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Remote:", conn.RemoteAddr())
	fmt.Println("Local:", conn.LocalAddr().String())
	fmt.Println("Protocol:", conn.LocalAddr().Network())

	request, err := handleRequest(conn)
	if err != nil {
		fmt.Println("Error handling request: ", err.Error())
		os.Exit(1)
	}

	err = handleResponse(conn, request)

	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}

	err = conn.Close()
	if err != nil {
		fmt.Println("Error closing connection: ", err.Error())
		os.Exit(1)
	}
}

func main() {
	// Create Listener
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
