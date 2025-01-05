package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	serverAddress = "localhost:8080"
)

func main() {
	fmt.Println("Starting Tic Tac Toe client...")
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Failed to connect to the server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to the Tic Tac Toe server!")
	reader := bufio.NewReader(conn)

	initialMessage, err := reader.ReadString(':')
	if err != nil {
		fmt.Println("Error receiving initial message:", err)
		return
	}
	fmt.Print(initialMessage + " ")

	// Start a goroutine to continuously read messages from the server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			message := scanner.Text()
			fmt.Println(message)
			if strings.Contains(message, "Game over!") {
				os.Exit(0)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Connection closed:", err)
			os.Exit(1)
		}
	}()

	// Send user input to the server
	consoleScanner := bufio.NewScanner(os.Stdin)
	for consoleScanner.Scan() {
		text := consoleScanner.Text()
		if _, err := fmt.Fprintln(conn, text); err != nil {
			fmt.Println("Failed to send input to the server:", err)
			return
		}
	}

	if err := consoleScanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
