package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func getFile(folderPath string, fileName string) ([]byte, error) {
	path := folderPath + "/" + fileName
	fmt.Println("looking for", path)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(content))
	return content, nil

}

func writeFile(folderPath string, fileName string, content []byte) error {
	path := folderPath + "/" + fileName
	fmt.Println("Trying to write to", path)
	err := os.WriteFile(path, content, 0664)
	return err
}

func main() {

	// folderPath := strings.Split(os.Args[1], "--directory")[1]
	// fmt.Println("folder path is", folderPath)

	fmt.Println("Logs from your program will appear here!")

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

	method := splits[0]
	path := splits[1]
	headers := extractHeaders(reqString)

	splits = strings.Split(path, "/")
	if path == "/" {
		con.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	}
	if strings.HasPrefix(path, "/files") {
		// folderPath := strings.Split(os.Args[1], "--directory")[1]
		fileName := strings.Split(path, "/")[2]
		folderPath := os.Args[2]
		if method == "POST" {
			body, err := extractBody(reqString, headers)
			if err != nil {
				fmt.Println("could not extract body", err)
				os.Exit(1)
			}
			err2 := writeFile(folderPath, fileName, body)
			if err2 != nil {
				fmt.Println("could not write file", err2)
				os.Exit(1)
			}

			con.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))

		} else {

			fmt.Println(os.Args)

			fmt.Println("DIR IS", folderPath, "FILE NAME IS", fileName)
			file, err := getFile(folderPath, fileName)
			fmt.Println("file is ", string(file))
			length := len(string(file))
			if err == nil {
				reqString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", length, string(file))
				con.Write([]byte(reqString))
			} else {
				fmt.Println("error is", err)
				con.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}
		}
		return
	}

	if path == "/user-agent" {
		userAgent := headers["user-agent"]
		resString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		con.Write([]byte(resString))
		return
	}

	if len(splits) == 3 && splits[1] == "echo" {

		echo := splits[2]
		encodings := strings.Split(headers["accept-encoding"], ", ")
		for _, encoding := range encodings {
			if encoding == "gzip" {
				zippedEcho, err := gzipCompress(echo)
				if err != nil {
					fmt.Println("cant compress", err)
					os.Exit(1)
				}
				resString := "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n%s"
				con.Write([]byte(fmt.Sprintf(resString, len(zippedEcho), zippedEcho)))
				return
			}
		}

		resString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		con.Write([]byte(resString))

		return
	}

	con.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

}

func gzipCompress(input string) (string, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	defer writer.Close()

	_, err := writer.Write([]byte(input))
	if err != nil {
		return "", err
	}
	compressed := buffer.String()
	return compressed, nil

}

func extractHeaders(reqString string) map[string]string {
	splits := strings.Split(reqString, "\r\n")
	fmt.Println("splits", splits, len(splits))
	headers := make(map[string]string)
	for i := 1; i < len(splits)-2; i++ {
		headerString := splits[i]
		keyAndValue := strings.Split(headerString, ": ")

		headers[strings.ToLower(keyAndValue[0])] = keyAndValue[1]
	}

	return headers

}

func extractBody(reqString string, headers map[string]string) ([]byte, error) {
	splits := strings.Split(reqString, "\r\n")
	body := splits[len(splits)-1]
	contentLength, err := strconv.Atoi(headers["content-length"])
	if err != nil {
		return []byte(""), err
	}
	body = body[:contentLength]
	return []byte(body), nil
}
