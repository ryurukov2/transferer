package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

const (
	udpPort = ":9999"
	tcpPort = ":8888"
)

// Starts a UDP server at port 9999 on the local machine
func startUDPServer() {
	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0"+udpPort)
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to listen on UDP port:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP server listening on", udpPort)

	buf := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}
		message := string(buf[:n])
		if message == "DISCOVER_FILE_SERVER" {
			response := fmt.Sprintf("FILE_SERVER_RESPONSE:%s", tcpPort)
			conn.WriteToUDP([]byte(response), clientAddr)
			fmt.Println("Sent discovery response to", clientAddr)
		}
	}
}

// Starts a TCP server at port 8888 on the local machine
func startTCPServer() {
	listener, err := net.Listen("tcp", tcpPort)
	if err != nil {
		fmt.Println("Failed to start TCP server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("TCP server listening on", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		go handleTCPConnection(conn)
	}
}

// Handles TCP connection for sending a file.
func handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from TCP connection:", err)
		return
	}

	filename := string(buf[:n])
	fmt.Println("Client requested file:", filename)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		conn.Write([]byte("ERROR: File not found"))
		return
	}
	defer file.Close()

	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file:", err)
		return
	}

	fmt.Println("File sent successfully")
}
