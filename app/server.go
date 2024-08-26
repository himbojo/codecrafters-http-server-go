package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type StatusLine struct {
	HTTPVersion          string
	StatusCode           int
	OptionalReasonPhrase string
}

func NewStatusLine(httpVersion string, statusCode int, optionalReponsePhrase string) StatusLine {
	statusLine := StatusLine{
		HTTPVersion:          "HTTP/1.1",
		StatusCode:           200,
		OptionalReasonPhrase: "",
	}

	if httpVersion != "" {
		statusLine.HTTPVersion = httpVersion
	}

	if statusCode != 0 {
		statusLine.StatusCode = statusCode
	}

	if optionalReponsePhrase != "" {
		statusLine.OptionalReasonPhrase = optionalReponsePhrase
	}

	return statusLine
}

func (statusLine StatusLine) ToString() string {
	if statusLine.OptionalReasonPhrase != "" {
		return strings.Join([]string{statusLine.HTTPVersion, strconv.Itoa(statusLine.StatusCode), statusLine.OptionalReasonPhrase}, " ") + "\r\n"
	} else {
		return strings.Join([]string{statusLine.HTTPVersion, strconv.Itoa(statusLine.StatusCode)}, " ") + "\r\n"
	}

}

func handleConnection(conn net.Conn) {
	fmt.Println("Remote:", conn.RemoteAddr(), "to Local:", conn.LocalAddr().String(), "via:", conn.LocalAddr().Network())
	// Status Line
	status := NewStatusLine("", 0, "").ToString()

	// Headers
	headers := "\r\n"

	// Respons Body
	body := ""

	httpResponse := strings.Join([]string{status, headers, body}, "")
	fmt.Printf("%q\n", httpResponse)

	httpResponseBytes := []byte(httpResponse)
	_, err := conn.Write(httpResponseBytes)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}

	// err = conn.Close()
	// if err != nil {
	// 	fmt.Println("Error closing connection: ", err.Error())
	// 	os.Exit(1)
	// }

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
