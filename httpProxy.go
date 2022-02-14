package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

// reads until stop or eof
func readConn (conn net.Conn) ([]byte, error) {
	var buf []byte // big buffer
	for {
		tmp := make([]byte, 256)
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
				return nil, err
			}
			break
		}
		buf = append(buf, tmp[:n]...)
		if n < len(tmp) {
			break
		}
	}
	log.Println("total size:", len(buf))
	return buf, nil
}

// parse and refactor request
func parseRequest (req []byte) (string, []byte, error) {
	split := strings.Split(string(req), "\n")
	addr := strings.Split(split[0], " ")
	urlString := addr[1]
	URL, err := url.Parse(urlString)
	if err != nil {
		return "", nil, err
	}
	host := URL.Host
	if URL.Port() == "" {
		host += ":80"
	}
	path := URL.Path
	// replace URL + delete headers
	edited := strings.Replace(string(req), urlString, path, 1)
	start := strings.Index(edited, "Proxy-Connection")
	end := strings.Index(edited[start:], "\n") + start + 1
	edited = edited[:start] + edited[end:]
	return host, []byte(edited), nil
}

func handle(conn net.Conn) {
	log.Print("Serving " + conn.RemoteAddr().String())
	defer conn.Close()
	//read rq
	reqUnparsed, err := readConn(conn)
	if err != nil {
		log.Fatal("wrong request")
		return
	}
	// parse + edit
	host, rq, err := parseRequest(reqUnparsed)
	if err != nil {
		log.Fatal("could not parse url")
		return
	}
	// conn to proxy
	proxyConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Fatal("could not connect to host")
		return
	}
	defer proxyConn.Close()
	// send to host
	_, err = proxyConn.Write(rq)
	if err != nil {
		log.Fatalf("error writing to host: %v\n", err)
		return
	}
	// read from host
	resp, err := readConn(proxyConn)
	if err != nil {
		log.Fatal("could not read response")
		return
	}
	// respond to request
	_, err = conn.Write(resp)
	if err != nil {
		log.Fatalf("error responding: %v\n", err)
		return
	}
	return
}

func main() {
	ln, err := net.Listen("tcp", ":6060")
	if err != nil {
		log.Fatal("failed to listen")
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("failed to accept")
			continue
		}
		go handle(conn)
	}
}