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
	err := os.WriteFile(os.Args[2]+filename, []byte(data), 0644)
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

	if request.Headers["Accept-Encoding"] == "gzip" {
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
				response.Headers["Content-Type"] = "text/octet-stream"
				response.Headers["Content-Length"] = fmt.Sprintf("%d", len(file))
				response.Body = file
			}
		default:
			response.StatusCode = 404
			response.Reason = "Not Found"
		}
	case "POST":
		switch request.Target {
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
	conn.Write(buildResponse(response))

	//// Request Regex
	//reqRegexp, _ := regexp.Compile(`^(?P<method>[A-Z]+) /(?P<targets>[^ ]+)? (?P<version>HTTP/[0-9.]+)`)
	//
	//// Parse Request
	//matches := reqRegexp.FindStringSubmatch(string(res))
	//
	//methodIndex := reqRegexp.SubexpIndex("method")
	//method := matches[methodIndex]
	//
	//targetIndex := reqRegexp.SubexpIndex("targets")
	//targets := matches[targetIndex]
	//
	//resArr := strings.Split(string(res), "\r\n")
	//body := resArr[len(resArr)-1]
	//
	//// Header Regex
	//agentRegexp, _ := regexp.Compile(`User-Agent: (?P<useragent>\S+)`)
	//accptEncRegexp, _ := regexp.Compile(`Accept-Encoding: (?P<encoding>\S+)`)
	//
	//// Parse Headers
	//agentMatches := agentRegexp.FindStringSubmatch(string(res))
	//agentIndex := agentRegexp.SubexpIndex("useragent")
	//userAgent := ""
	//if len(agentMatches) != 0 {
	//	userAgent = agentMatches[agentIndex]
	//}
	//
	//accptEncMatches := accptEncRegexp.FindStringSubmatch(string(res))
	//accptEncIndex := accptEncRegexp.SubexpIndex("encoding")
	//acceptEncoding := ""
	//if len(accptEncMatches) != 0 {
	//	acceptEncoding = accptEncMatches[accptEncIndex]
	//}
	//
	//// Route Regex
	//echoRegexp, _ := regexp.Compile(`^echo/(?P<echo>[^ /]+)$`)
	//userAgentRegexp, _ := regexp.Compile(`^user-agentRegexp$`)
	//filesRegexp, _ := regexp.Compile(`^files/(?P<filename>[^ ]+)$`)
	//
	//// Response
	//out := "HTTP/1.1 404 Not Found\r\n\r\n"
	//
	//switch method {
	//case "GET":
	//	switch {
	//	case len(targets) == 0:
	//		out = "HTTP/1.1 200 OK\r\n\r\n"
	//	case acceptEncoding == "gzip":
	//		out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\n\r\n%s", body)
	//	case echoRegexp.MatchString(targets):
	//		echoMatches := echoRegexp.FindStringSubmatch(targets)
	//		echoIndex := echoRegexp.SubexpIndex("echo")
	//		echo := echoMatches[echoIndex]
	//		out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	//	case userAgentRegexp.MatchString(targets):
	//		out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	//	case filesRegexp.MatchString(targets):
	//		fileMatches := filesRegexp.FindStringSubmatch(targets)
	//		if len(fileMatches) != 0 {
	//			filenameIndex := filesRegexp.SubexpIndex("filename")
	//			filename := fileMatches[filenameIndex]
	//			data, err := os.ReadFile(os.Args[2] + filename)
	//
	//			if err != nil {
	//				fmt.Println("Error reading file: " + err.Error())
	//			} else {
	//				sData := string(data)
	//				out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(sData), sData)
	//			}
	//		}
	//	}
	//case "POST":
	//	switch {
	//	case filesRegexp.MatchString(targets):
	//		fileMatches := filesRegexp.FindStringSubmatch(targets)
	//		if len(fileMatches) != 0 {
	//			filenameIndex := filesRegexp.SubexpIndex("filename")
	//			filename := fileMatches[filenameIndex]
	//			fmt.Println(os.Args[2]+filename, body)
	//			file, err := os.OpenFile(os.Args[2]+filename, os.O_WRONLY|os.O_CREATE, 0644)
	//			if err != nil {
	//				fmt.Println("Error reading file: " + err.Error())
	//			} else {
	//				defer file.Close()
	//				file.Write([]byte(strings.Trim(body, "\x00")))
	//				out = "HTTP/1.1 201 Created\r\n\r\n"
	//			}
	//		}
	//	}
	//}
	//conn.Write([]byte(out))
	//
	//fmt.Println("----------------")
	//fmt.Println("Sent:")
	//fmt.Println(out)
	conn.Close()
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
