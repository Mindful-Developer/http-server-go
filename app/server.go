package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
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

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	res := make([]byte, 1024)
	_, err = conn.Read(res)
	if err != nil {
		fmt.Println("Error reading from connected client: ", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(res))

	r, _ := regexp.Compile("^(?P<method>[A-Z]+) (?P<target>[^ ]+) (?P<version>HTTP/[0-9.]+)")
	matches := r.FindStringSubmatch(string(res))
	targetIndex := r.SubexpIndex("target")
	target := matches[targetIndex]

	// Route matching
	echoRegexp, _ := regexp.Compile("^/echo/(?P<echo>[^ /]+)$")

	out := ""
	if target == "/" {
		out = "HTTP/1.1 200 OK\r\n\r\n"
	} else if echoRegexp.MatchString(target) {
		echoMatches := echoRegexp.FindStringSubmatch(target)
		echoIndex := echoRegexp.SubexpIndex("echo")
		echo := echoMatches[echoIndex]
		out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	} else {
		out = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(out))

	fmt.Println(out)
}
