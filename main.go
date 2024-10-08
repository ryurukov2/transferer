package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select mode: (1) Start Server (2) Scan for Servers")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		go startUDPServer()
		startTCPServer()
	case "2":
		serverAddrs, err := discoverServers()
		if err != nil {
			fmt.Println("Server discovery failed:", err)
			return
		}

		fmt.Println("Discovered servers at:", serverAddrs)
		// err = requestFile(serverAddr, "example.txt")
		if err != nil {
			fmt.Println("File request failed:", err)
		}
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}

}
