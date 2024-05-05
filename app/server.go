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
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 1024)
	total, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		return
	}

	requestLines := strings.Split(string(buff)[:total], "\r\n")

	startLine := requestLines[0]

	startLineSeparated := strings.Split(startLine, " ")

	method := startLineSeparated[0]
	path := startLineSeparated[1]

	if path == "/" {
		writeResponse(conn, 200, "", "")
		return
	}

	if strings.HasPrefix(path, "/files/") {
		fmt.Println("debug")

		fileName, found := strings.CutPrefix(path, "/files/")

		if !found || len(os.Args) < 3 {
			writeResponse(conn, 404, "", "")
			return
		}

		directoryPath := os.Args[2]

		if method == "POST" {

			if err := os.WriteFile(directoryPath+fileName, []byte(requestLines[len(requestLines)-1]), 066); err != nil {
				writeResponse(conn, 500, "", "")
				return
			}

			writeResponse(conn, 201, "", "")
			return
		} else {
			content, err := os.ReadFile(directoryPath + fileName)
			if err != nil {
				writeResponse(conn, 404, "", "")
				return
			}

			writeResponse(conn, 200, string(content), "application/octet-stream")
			return
		}
	}

	if path == "/user-agent" {
		userAgentLine := requestLines[2]
		userAgent := strings.Split(userAgentLine, " ")[1]

		writeResponse(conn, 200, userAgent, "text/plain")
		return
	}

	substr := "/echo/"
	if !strings.HasPrefix(path, substr) {
		writeResponse(conn, 404, "", "")
		return
	}

	substrIndex := strings.LastIndex(path, substr)
	stringFromPath := path[len(substr)+substrIndex:]

	writeResponse(conn, 200, stringFromPath, "text/plain")
}

func writeResponse(conn net.Conn, status int, body, contentType string) {
	responseText := ""

	switch status {
	case 200:
		responseText += "HTTP/1.1 200 OK"
	case 201:
		responseText += "HTTP/1.1 201 Created"
	case 404:
		responseText += "HTTP/1.1 404 Not Found"
	default:
		responseText += "HTTP/1.1 500 Internal Error"
	}

	if body != "" && contentType != "" {
		responseText +=
			"\r\nContent-Type: " + contentType + "\r\nContent-Length: " +
				fmt.Sprint(len(body)) + "\r\n\r\n" + body
	} else {
		responseText += "\r\n\r\n"
	}

	_, err := conn.Write([]byte(responseText))

	if err != nil {
		fmt.Print("Error writing response: ", err.Error())
	}
}
