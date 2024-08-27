package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type ResponseStatusLine struct {
	HTTPVersion          string
	StatusCode           int
	OptionalReasonPhrase string
}

func (statusLine ResponseStatusLine) ToString() string {
	return fmt.Sprintf("%s %d %s\r\n", statusLine.HTTPVersion, statusLine.StatusCode, statusLine.OptionalReasonPhrase)
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

func handleRequest(conn net.Conn) (Request, error) {
	reader := bufio.NewReader(conn)

	// Read the request line
	requestLineString, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, fmt.Errorf("error reading request line: %w", err)
	}

	requestLineArray := strings.Fields(requestLineString)
	if len(requestLineArray) < 3 {
		return Request{}, fmt.Errorf("malformed request line")
	}
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
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return Request{}, fmt.Errorf("malformed header line")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headerMap[key] = value
	}

	fmt.Println("Headers:")
	for key, value := range headerMap {
		fmt.Printf("%s: %s\n", key, value)
	}

	// Read the rest of the data into the body
	// body, err := reader.ReadBytes('\n')
	// if err != nil {
	// 	return Request{}, fmt.Errorf("error reading body: %w", err)
	// }
	var body []byte
	if requestLine.HTTPMethod == "POST" {
		contentLength, err := strconv.Atoi(headerMap["Content-Length"])
		if err != nil {
			fmt.Println("Error:", err)
			return Request{}, fmt.Errorf("error converting Content-Length: %w", err)
		}
		buffer := make([]byte, contentLength)
		_, err = reader.Read(buffer)
		if err != nil {
			return Request{}, fmt.Errorf("error reading bytes: %w", err)
		}
		body = append(body, buffer...)
	}

	fmt.Println("\nBody:")
	fmt.Print(string(body))

	request := Request{
		StatusLine: requestLine,
		Headers:    headerMap,
		Body:       body,
	}

	return request, nil
}

func fileExistsInDirectory(directory, filename string) (bool, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if file.Name() == filename {
			return true, nil
		}
	}
	return false, nil
}

func readFileIntoByteArray(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
		// buffer.WriteByte('\n') // Add newline character to preserve line breaks
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return buffer.Bytes(), nil
}

func handleResponse(conn net.Conn, request Request, directory string) error {
	responseStatusLine := ResponseStatusLine{
		HTTPVersion:          "HTTP/1.1",
		StatusCode:           200,
		OptionalReasonPhrase: "OK",
	}

	headers, responseBody, err := generateResponse(request, directory, &responseStatusLine)
	if err != nil {
		return err
	}

	headers += "\r\n"
	httpResponse := fmt.Sprintf("%s%s%s", responseStatusLine.ToString(), headers, responseBody)
	fmt.Printf("\nResponse:\n%q\n", httpResponse)

	if _, err := conn.Write([]byte(httpResponse)); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}
	return nil
}

func generateResponse(request Request, directory string, responseStatusLine *ResponseStatusLine) (string, string, error) {
	headers := ""
	responseBody := ""

	switch {
	case request.StatusLine.RequestTarget == "/":
		// No additional headers or body
	case strings.HasPrefix(request.StatusLine.RequestTarget, "/echo"):
		return handleEcho(request, &headers, &responseBody)
	case strings.HasPrefix(request.StatusLine.RequestTarget, "/user-agent"):
		return handleUserAgent(request, &headers, &responseBody)
	case strings.HasPrefix(request.StatusLine.RequestTarget, "/files") && request.StatusLine.HTTPMethod == "GET":
		return handleFileGet(request, directory, responseStatusLine, &headers, &responseBody)
	case strings.HasPrefix(request.StatusLine.RequestTarget, "/files") && request.StatusLine.HTTPMethod == "POST":
		return handleFilePost(request, directory, responseStatusLine)
	default:
		responseStatusLine.StatusCode = 404
		responseStatusLine.OptionalReasonPhrase = "Not Found"
	}

	return headers, responseBody, nil
}

func handleEncoding(request Request, headers string) string {
	acceptEncoding, exists := request.Headers["Accept-Encoding"]
	if exists {
		encodingTypes := strings.Split(acceptEncoding, ",")
		for _, encodingType := range encodingTypes {
			fmt.Printf("\nClients supported encoding:\n%s\n", encodingType)
			if strings.TrimSpace(encodingType) == "gzip" {
				headers += "Content-Encoding: gzip\r\n"
			}
		}
	}
	fmt.Printf("\nCurrently Headers looks like:\n%s", headers)
	return headers

}

func handleEcho(request Request, headers, responseBody *string) (string, string, error) {
	segments := strings.Split(request.StatusLine.RequestTarget, "/")
	if len(segments) != 3 {
		return "", "", fmt.Errorf("incorrect endpoint: Expected %s/{STR}", request.StatusLine.RequestTarget)
	}

	*headers += "Content-Type: text/plain\r\n"
	*headers = handleEncoding(request, *headers)
	*headers += fmt.Sprintf("Content-Length: %d\r\n", len(segments[2]))
	fmt.Printf("\nNow Headers looks like this:\n%s\n", *headers)
	*responseBody += segments[2]
	return *headers, *responseBody, nil
}

func handleUserAgent(request Request, headers, responseBody *string) (string, string, error) {
	userAgent := request.Headers["User-Agent"]
	*headers += "Content-Type: text/plain\r\n"
	*headers += fmt.Sprintf("Content-Length: %d\r\n", len(userAgent))
	*responseBody += userAgent
	return *headers, *responseBody, nil
}

func handleFileGet(request Request, directory string, responseStatusLine *ResponseStatusLine, headers, responseBody *string) (string, string, error) {
	segments := strings.Split(request.StatusLine.RequestTarget, "/")
	if len(segments) != 3 {
		return "", "", fmt.Errorf("incorrect endpoint: Expected %s/{filename}", request.StatusLine.RequestTarget)
	}
	filename := segments[2]

	exists, err := fileExistsInDirectory(directory, filename)
	if err != nil {
		return "", "", fmt.Errorf("error checking if file exists: %w", err)
	}
	if !exists {
		responseStatusLine.StatusCode = 404
		responseStatusLine.OptionalReasonPhrase = "Not Found"
		return *headers, *responseBody, nil
	}

	content, err := readFileIntoByteArray(fmt.Sprintf("%s%s", directory, filename))
	if err != nil {
		return "", "", fmt.Errorf("error reading file: %w", err)
	}

	*headers += "Content-Type: application/octet-stream\r\n"
	*headers += fmt.Sprintf("Content-Length: %d\r\n", len(content))
	*responseBody += string(content)
	return *headers, *responseBody, nil
}

func handleFilePost(request Request, directory string, responseStatusLine *ResponseStatusLine) (string, string, error) {
	segments := strings.Split(request.StatusLine.RequestTarget, "/")
	if len(segments) != 3 {
		return "", "", fmt.Errorf("incorrect endpoint: Expected %s/{filename}", request.StatusLine.RequestTarget)
	}
	filename := segments[2]

	content := request.Body
	file, err := os.Create(fmt.Sprintf("%s/%s", directory, filename))
	if err != nil {
		return "", "", fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	if _, err = file.Write(content); err != nil {
		return "", "", fmt.Errorf("error writing file: %w", err)
	}
	responseStatusLine.StatusCode = 201
	responseStatusLine.OptionalReasonPhrase = "Created"
	return "", "", nil
}

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()
	fmt.Println("Remote:", conn.RemoteAddr())
	fmt.Println("Local:", conn.LocalAddr().String())
	fmt.Println("Protocol:", conn.LocalAddr().Network())

	request, err := handleRequest(conn)
	if err != nil {
		fmt.Println("Error handling request: ", err.Error())
		os.Exit(1)
	}

	if err := handleResponse(conn, request, directory); err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func main() {
	// Create Listener
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer listener.Close()
	var directory string
	if len(os.Args) < 2 {
		fmt.Println("No command line arguments provided.")
	} else {
		for i, arg := range os.Args {
			switch arg {
			case "--directory":
				if i+1 < len(os.Args) {
					directory = os.Args[i+1]
				} else {
					fmt.Println("No directory specified after --directory")
					return
				}

				exists, err := directoryExists(directory)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				if exists {
					fmt.Printf("Directory %s exists.\n", directory)
				} else {
					fmt.Printf("Directory %s does not exist.\n", directory)
					return
				}
			}
		}
	}
	// this should already handle concurrent requests
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, directory)
	}
}
