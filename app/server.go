package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
	fmt.Println("----------------")
	fmt.Println("Received:")
	fmt.Println(string(res))

	// Request Regex
	req, _ := regexp.Compile("^(?P<method>[A-Z]+) /(?P<targets>[^ ]+)? (?P<version>HTTP/[0-9.]+)")
	agent, _ := regexp.Compile(`User-Agent: (?P<useragent>\S+)`)

	// Parse Request
	matches := req.FindStringSubmatch(string(res))
	targetIndex := req.SubexpIndex("targets")
	targets := matches[targetIndex]

	agentMatches := agent.FindStringSubmatch(string(res))
	agentIndex := agent.SubexpIndex("useragent")
	userAgent := ""
	if len(agentMatches) != 0 {
		userAgent = agentMatches[agentIndex]
	}

	targetArr := []string{}
	if targets != "" {
		targetArr = strings.Split(targets, "/")
	}

	// fmt.Println("----------------")
	// fmt.Println("Matches:")
	// fmt.Println(targets)
	// fmt.Println(len(targetArr))
	// for i, target := range targetArr {
	// 	fmt.Printf("%d: %s\n", i, target)
	// }

	// Route Regex
	echoRegexp, _ := regexp.Compile("^echo/(?P<echo>[^ /]+)$")
	userAgentRegexp, _ := regexp.Compile("^user-agent$")


	// Response
	out := ""
	if len(targetArr) == 0 {
		out = "HTTP/1.1 200 OK\r\n\r\n"
	} else if echoRegexp.MatchString(targets) {
		echoMatches := echoRegexp.FindStringSubmatch(targets)
		echoIndex := echoRegexp.SubexpIndex("echo")
		echo := echoMatches[echoIndex]
		out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	} else if userAgentRegexp.MatchString(targets) {
		out = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	} else {
		out = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(out))

	fmt.Println(out)
}
