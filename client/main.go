package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"proof_of_work/pkg"
	"strconv"
	"strings"
	"time"
)

func solveChallenge(message string) (string, error) {
	key, leadingZeros, err := parseChallenge(message)
	if err != nil {
		return "", err
	}
	hash, nonce, err := pkg.CalculateHashWithLeadingZeros(key, 0, leadingZeros)
	if err != nil {
		return "", err
	}

	message = fmt.Sprintf("%s:%v", hash, nonce)
	return message, nil
}

func parseChallenge(message string) (int64, int, error) {
	challenges := strings.Split(message, ":")
	if len(challenges) != 2 {
		return 0, 0, errors.New("ERROR: wrong count of parts")
	}

	keyInt, err := strconv.ParseInt(challenges[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("ERROR: wrong type of key")
	}

	leadingZerosInt, err := strconv.Atoi(challenges[1])
	if err != nil {
		return 0, 0, errors.New("ERROR: wrong type of leading zeroes")
	}

	return keyInt, leadingZerosInt, nil
}

func main() {
	hostFlag := flag.String("host", "localhost:8080", "host of server")
	timeoutFlag := flag.Duration("timeout", 15*time.Second, "set timeout for connection")
	flag.Parse()

	timeout := *timeoutFlag
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Connect to the server
	conn, err := dialer.Dial("tcp", *hostFlag)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil && errors.Is(err, net.ErrClosed) {
		fmt.Println("Error reading:", err)
		return
	}
	response := string(buffer[:n])

	answer, err := solveChallenge(response)
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}
	conn.Write([]byte(answer))

	// Read response from the server
	buffer = make([]byte, 1024)
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	response = string(buffer[:n])
	if len(response) == 0 {
		fmt.Println("EOF")
		return
	}
	fmt.Println("Got message:", response)
}
