package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// -- Types --

// Represents the results from checking a URL.
type Answer struct {
	name   string
	result bool
}

// -- Functions --

func canConnect(url string) bool {
	resp, err := http.Get(url)

	if err != nil {
		return false
	}

	resp.Body.Close()

	if resp.StatusCode > 399 {
		return false
	} else {
		return true
	}
}

func testURL(name, url string, chRes chan Answer, chDone chan bool) {
	res := Answer{name, canConnect(url)}

	chRes <- res
	chDone <- true
}

func parseConfig(source io.ReadCloser) (map[string]string, error) {
	// When parsing the config each test should be on it's own line in the
	// form 'name = url' (whitespace around name and url is ignored). If
	// there are duplicate names then the last one that is found will be
	// used.
	//
	// Any lines that start with '#' are treated as comment.
	//
	// All of the following are valid lines and are considered equal:
	// Example = https://www.example.com
	//  Example = https://www.example.com
	// Example=https://www.example.com
	// Example =https://www.example.com
	// Example= https://www.example.com
	//
	// Lines that don't match that pattern are silently ignored. The
	// following lines would be ignored:
	// Example
	// Example = https://www.example.com = asd
	urls := make(map[string]string)

	scanner := bufio.NewScanner(source)

	for scanner.Scan() {
		line := scanner.Text()

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			url := strings.TrimSpace(parts[1])

			urls[name] = url
		}
	}
	return urls, scanner.Err()
}

func getConfigFile() (io.ReadCloser, error) {
	// Try to find our config file in the users home .config dir e.g
	// ~/.config/is_it_down.
	user, err := user.Current()

	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(user.HomeDir, ".config", "is_it_down")

	file, err := os.Open(configFile)

	if err != nil {
		return nil, err
	}

	return file, nil
}

// -- Main --

func main() {
	chRes := make(chan Answer)
	chDone := make(chan bool)

	// TODO - At some point it might be worth replacing this with a better
	// TODO - Better error handling

	file, err := getConfigFile()
	defer file.Close()

	if err != nil {
		fmt.Println("Error opening config")
		fmt.Println(err)
		os.Exit(1)
	}

	urls, err := parseConfig(file)
	if err != nil {
		fmt.Println("Error parsing config")
		fmt.Println(err)
		os.Exit(1)
	}

	for name, url := range urls {
		go testURL(name, url, chRes, chDone)
	}

	for c := 0; c < len(urls); {
		select {
		case answer := <-chRes:
			if answer.result {
				fmt.Printf("%s is all ok :D\n", answer.name)
			} else {
				fmt.Printf("%s is Down :(\n", answer.name)
			}
		case <-chDone:
			c++
		}
	}
}
