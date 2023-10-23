package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"proof_of_work/pkg"
)

var count atomic.Int64

func handleConnection(
	conn net.Conn,
	challenger *pkg.Challenger,
	wg *sync.WaitGroup,
	interrupt chan bool) {
	defer func() {
		conn.Close()
		wg.Done()
	}()

	select {
	case <-interrupt:
		conn.Close()
	default:
		localCount := count.Add(1)
		fmt.Println("Have new connection ", localCount)

		buffer := make([]byte, 1024)
		key, leadingZeros := challenger.GenerateChallenge()

		message := fmt.Sprintf("%v:%v", key, leadingZeros)

		conn.Write([]byte(message))
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				return
			}

			data := buffer[:n]
			body := strings.Split(string(data), ":")
			if len(body) != 2 {
				conn.Write([]byte("ERROR: wrong len of answer"))
				return
			}

			hashInAnswer := body[0]
			nonce := body[1]
			switch body[0] {
			case "ERROR":
				fmt.Println("Got error:", body)
				return
			default:
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

				if hash != hashInAnswer {
					conn.Write([]byte("ERROR: wrong check"))
					return
				}

				idx := rand.Intn(len(pkg.Quotes) - 1)

				conn.Write([]byte(pkg.Quotes[idx]))
				fmt.Println("Message sent to connection", localCount)
				return
			}
		}
	}
}

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		fmt.Println("Empty host env")
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Empty port env")
		return
	}

	timeout := os.Getenv("TIMEOUT")
	if timeout == "" {
		fmt.Println("Empty timeout env")
	}

	leadingZerosStr := os.Getenv("LEADING_ZEROS")
	if leadingZerosStr == "" {
		fmt.Println("Empty leading zeros env")
	}

	randomRangeStr := os.Getenv("RANDOM_RANGE")
	if leadingZerosStr == "" {
		fmt.Println("Empty random range env")
	}

	var addr = fmt.Sprintf("%s:%s", host, port)

	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		fmt.Println("Error parsing timeout:", err)
		return
	}

	leadingZeros, err := strconv.Atoi(leadingZerosStr)
	if err != nil {
		fmt.Println("Error parsing leading zeros:", err)
		return
	}

	randomRange, err := strconv.ParseInt(randomRangeStr, 10, 64)
	if err != nil {
		fmt.Println("Error parsing random range:", err)
		return
	}

	// Graceful shutdown channel for goroutines
	interrupt := make(chan bool)
	var wg sync.WaitGroup

	// Set up signal handling to interrupt the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	// Graceful shutdown
	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal %s. Shutting down...\n", sig)
		close(interrupt)
		listener.Close()
	}()

	fmt.Println("Server started. Listening on address ", addr)

	challenger := pkg.NewChallenger(leadingZeros, randomRange)
	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		wg.Add(1)
		conn.SetDeadline(time.Now().Add(timeoutDur))
		// Handle the connection in a separate goroutine
		go handleConnection(conn, challenger, &wg, interrupt)
	}
	wg.Wait()
	fmt.Println("Server stopped")
}
