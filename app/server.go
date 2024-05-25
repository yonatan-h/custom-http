package main

import (
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(con)
	}
}

func handleConnection(con net.Conn) {

	readBuffer := make([]byte, 1000)
	_, err2 := con.Read(readBuffer)
	if err2 != nil {
		fmt.Println("Error reading req", err2.Error())
		os.Exit(1)
	}
	reqString := string(readBuffer)
	fmt.Println("req string", reqString)
	splits := strings.Split(reqString, "\r\n") //Get /app HTTP/1.1 ...headers
	splits = strings.Split(splits[0], " ")

	path := splits[1]
	headers := extractHeaders(reqString)

	splits = strings.Split(path, "/")
	if path == "/" {
		con.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	}

	if path == "/user-agent" {
		userAgent := headers["User-Agent"]
		resString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		con.Write([]byte(resString))
		return
	}

	if len(splits) != 3 || splits[1] != "echo" {
		con.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	echo := splits[2]
	resString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	con.Write([]byte(resString))
}

func extractHeaders(reqString string) map[string]string {
	splits := strings.Split(reqString, "\r\n")
	fmt.Println("splits", splits, len(splits))
	headers := make(map[string]string)
	for i := 1; i < len(splits)-2; i++ {
		headerString := splits[i]
		keyAndValue := strings.Split(headerString, ": ")
		headers[keyAndValue[0]] = keyAndValue[1]
	}

	return headers

}
