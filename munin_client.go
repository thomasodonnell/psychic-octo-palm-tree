package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func connect(protocall, addr string) (net.Conn, error) {
	conn, err := net.Dial(protocall, addr)

	if err == nil {
		// Discard the hello message
		bufio.NewReader(conn).ReadString('\n')
	}
	return conn, err
}

func list(conn net.Conn) ([]string, error) {
	fmt.Fprintf(conn, "list\n")

	raw, err := bufio.NewReader(conn).ReadString('\n')
	raw = strings.TrimSpace(raw)

	return strings.Split(raw, " "), err
}

func fetch(conn net.Conn, tag string) map[string]string {
	fmt.Fprintf(conn, "fetch %s \n", tag)
	scanner := bufio.NewScanner(conn)

	res := make(map[string]string)

	for scanner.Scan() {
		mes := scanner.Text()
		if mes == "." {
			break
		}
		split := strings.Split(mes, " ")
		res[split[0]] = split[1]
	}

	return res
}

func main() {
	conn, err := connect("tcp", "localhost:4949")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	metrics, _ := list(conn)
	//metrics := []string{"processes", "df", "uptime"}

	for _, metric := range metrics {
		for name, value := range fetch(conn, metric) {
			log.Printf("%s.%s : %s\n", metric, name, value)
		}
	}
}
