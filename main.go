package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type URL struct {
	Scheme string
	URL    string
	Host   string
	Path   string
}

func (u *URL) Init(url string) error {
	if len(url) == 0 {
		return errors.New("no url provided")
	}

	parts := strings.SplitN(url, "://", 2)
	if len(parts) > 2 || parts[0] != "http" {
		return errors.New("malformed url")
	}

	u.Scheme = parts[0]
	u.URL = parts[1]

	if !strings.Contains(u.URL, "/") {
		u.URL += "/"
	}

	parts = strings.SplitN(u.URL, "/", 2)

	u.Host = parts[0]
	u.URL = parts[1]
	u.Path = "/" + u.URL

	return nil
}

func (u *URL) Request() (string, error) {
	var err error
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:80", u.Host))

	if err != nil {
		return "", err
	}
	defer conn.Close()

	httpRequest := fmt.Sprintf("GET %s HTTP/1.0\r\nHost: %s\r\n\r\n", u.Path, u.Host)

	_, err = conn.Write([]byte(httpRequest))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadBytes('\n')
	if err != nil {
		return "", err
	}

	var version, status, explanation string

	parts := strings.SplitN(string(statusLine), " ", 3)
	version = parts[0]
	status = parts[1]
	explanation = parts[2]

	fmt.Println(version, status, explanation)
	headers := make(map[string]string, 0)

	for {
		line, err := reader.ReadString('\n') // Read until newline
		if line == "\r\n" {
			break // End of headers
		}
		if err != nil {
			fmt.Println("Error reading line:", err)
			return "", err
		}

		var name, value string
		parts := strings.SplitN(line, ":", 2)
		name = parts[0]
		value = parts[1]

		headers[name] = strings.Trim(value, " ")
	}

	if _, ok := headers["transfer-encoding"]; ok {
		return "", errors.New("response in unexpected format")
	}

	if _, ok := headers["content-encoding"]; ok {
		return "", errors.New("response in unexpected format")
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func Show(body string) {
	inTag := false
	for _, char := range body {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			fmt.Printf("%c", char)
		}
	}
}

func load(u URL) error {
	body, err := u.Request()
	if err != nil {
		return err
	}

	Show(body)

	return nil
}

func main() {
	u := URL{}
	err := u.Init("http://example.org")
	if err != nil {
		log.Fatal(err)
	}
	err = load(u)
	if err != nil {
		log.Fatal(err)
	}
}
