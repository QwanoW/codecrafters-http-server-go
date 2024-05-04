package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buff := make([]byte, 1024)

	_, err := conn.Read(buff)

	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}

	startLine := strings.Split(string(buff), "\r\n")[0]
	path := strings.Split(startLine, " ")[1]

	if path == "/" {
		writeResponse(conn, 200)
	}

	substr := "/echo/"
	if !strings.Contains(path, substr) {
		writeResponse(conn, 404)
	}

	substrIndex := strings.LastIndex(path, substr)
	stringFromPath := path[len(substr)+substrIndex:]

	writeResponse(conn, 200, stringFromPath)
}

func writeResponse(conn net.Conn, status int, bodyOptional ...string) {
	responseText := ""

	body := ""
	if len(bodyOptional) > 0 {
		body = bodyOptional[0]
	}

	switch status {
	case 200:
		responseText += "HTTP/1.1 200 OK"
	case 404:
		responseText += "HTTP/1.1 404 Not Found"
	default:
		responseText += "HTTP/1.1 500 Internal Error"
	}

	if body != "" {
		responseText +=
			"\r\nContent-Type: text/plain\r\nContent-Length: " +
				fmt.Sprint(len(body)) + "\r\n\r\n" + body
	} else {
		responseText += "\r\n\r\n"
	}

	_, err := conn.Write([]byte(responseText))

	if err != nil {
		fmt.Print("Error writing response: ", err.Error())
	}
	
	os.Exit(1)
}
