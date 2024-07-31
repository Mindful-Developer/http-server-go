package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method  string
	Target  string
	Version string
	Headers map[string]string
	Body    string
}

type Response struct {
	StatusCode int
	Reason     string
	Headers    map[string]string
	Body       string
}

func buildResponse(response Response) []byte {
	headers := ""
	for key, value := range response.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", response.StatusCode, response.Reason, headers, response.Body))
}

func parseRequest(requestString string) Request {
	request := Request{}

	lines := strings.Split(requestString, "\r\n")
	requestLine := strings.Split(lines[0], " ")
	request.Method = requestLine[0]
	request.Target = requestLine[1]
	request.Version = requestLine[2]

	request.Headers = make(map[string]string)
	request.Body = ""
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			request.Body = strings.Join(lines[i+1:], "\r\n")
			break
		}
		header := strings.Split(lines[i], ": ")
		request.Headers[header[0]] = header[1]
	}
	return request
}

func downloadFile(filename string) (string, error) {
	if len(os.Args) < 3 {
		return "", fmt.Errorf("no directory provided")
	}
	file, err := os.ReadFile(os.Args[2] + filename)
	if err != nil {
		return "", err
	}
	return string(file), nil
}

func uploadFile(filename string, data string) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("no directory provided")
	}
	err := os.WriteFile(os.Args[2]+filename, []byte(strings.Trim(data, "\x00")), 0644)
	if err != nil {
		return err
	}
	return nil
}

func splitTarget(target string) (string, string) {
	parts := strings.Split(strings.Trim(target, "/"), "/")
	if len(parts) == 1 {
		return "/" + parts[0], ""
	} else {
		return "/" + parts[0], strings.Join(parts[1:], "/")
	}
}

func routeRequest(request Request) Response {
	response := Response{
		StatusCode: 404,
		Reason:     "Not Found",
		Headers:    make(map[string]string),
		Body:       "",
	}

	target, path := splitTarget(request.Target)

	encodings, ok := request.Headers["Accept-Encoding"]
	if ok && strings.Contains(encodings, "gzip") {
		response.Headers["Content-Encoding"] = "gzip"
	}

	switch request.Method {
	case "GET":
		switch target {
		case "/":
			response.StatusCode = 200
			response.Reason = "OK"
		case "/echo":
			response.StatusCode = 200
			response.Reason = "OK"
			response.Headers["Content-Type"] = "text/plain"
			response.Headers["Content-Length"] = fmt.Sprintf("%d", len(path))
			response.Body = path
		case "/user-agent":
			response.StatusCode = 200
			response.Reason = "OK"
			response.Headers["Content-Type"] = "text/plain"
			response.Headers["Content-Length"] = fmt.Sprintf("%d", len(request.Headers["User-Agent"]))
			response.Body = request.Headers["User-Agent"]
		case "/files":
			response.StatusCode = 200
			response.Reason = "OK"
			file, err := downloadFile(path)
			if err != nil {
				response.StatusCode = 404
				response.Reason = "Not Found"
			} else {
				response.Headers["Content-Type"] = "application/octet-stream"
				response.Headers["Content-Length"] = fmt.Sprintf("%d", len(file))
				response.Body = file
			}
		default:
			response.StatusCode = 404
			response.Reason = "Not Found"
		}
	case "POST":
		switch target {
		case "/files":
			response.StatusCode = 201
			response.Reason = "Created"
			err := uploadFile(path, request.Body)
			if err != nil {
				response.StatusCode = 406
				response.Reason = "Error uploading file"
			}
		default:
			response.StatusCode = 404
			response.Reason = "Not Found"
		}
	}
	return response
}

func handleConnection(conn net.Conn) {
	res := make([]byte, 1024)
	_, err := conn.Read(res)
	if err != nil {
		fmt.Println("Error reading from connected client: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("----------------")
	fmt.Println("Received:")
	fmt.Println(string(res))

	request := parseRequest(string(res))
	response := routeRequest(request)
	rStr := buildResponse(response)
	conn.Write(rStr)

	fmt.Println("----------------")
	fmt.Println("Sent:")
	fmt.Println(string(rStr))

	conn.Close()
}

func main() {
	var conn net.Conn
	var err error
	var l net.Listener

	l, err = net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err = l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
