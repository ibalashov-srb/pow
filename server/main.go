package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"proof_of_work/pkg"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var count atomic.Int64

func handleConnection(conn net.Conn) {
	defer conn.Close()
	localCount := count.Add(1)
	fmt.Println("Have new connection ", localCount)

	buffer := make([]byte, 1024)
	key, leadingZeros := pkg.GenerateChallenge()

	message := fmt.Sprintf("%v:%v", key, leadingZeros)

	conn.Write([]byte(message))
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		data := buffer[:n]
		body := strings.Split(string(data), ":")
		switch body[0] {
		case "ERROR":
			fmt.Println("Got error:", body)
			return
		default:
			nonce := body[1]
			nonceInt, err := strconv.ParseInt(nonce, 10, 64)
			if err != nil {
				fmt.Println("Can't parse nonce")
				return
			}

			hash, _, err := pkg.CalculateHashWithLeadingZeros(key, nonceInt, leadingZeros)
			if err != nil {
				fmt.Println("Can't find hash")
				return
			}

			if hash != body[0] {
				conn.Write([]byte("ERROR: wrong check"))
				return
			}

			idx := rand.Intn(len(pkg.Quotes) - 1)
			conn.Write([]byte(fmt.Sprintf("%s", pkg.Quotes[idx])))
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			fmt.Println("Message sent to connection", localCount)
			return
		}
	}
}

func main() {
	hostFlag := flag.String("host", "localhost:8080", "host of server")
	timeoutFlag := flag.Duration("timeout", 15*time.Second, "set timeout for connection")
	flag.Parse()

	// Start the server
	listener, err := net.Listen("tcp", *hostFlag)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started. Listening on port 8080...")

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		conn.SetDeadline(time.Now().Add(*timeoutFlag))
		// Handle the connection in a separate goroutine
		go handleConnection(conn)
	}
}
