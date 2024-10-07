package main

import (
	"fmt"
	"net"
)

const (
	udpPort = ":9999"
	tcpPort = ":8888"
)

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
		fmt.Printf("|%v|\n", message)
		if message == "DISCOVER_FILE_SERVER" {
			fmt.Println(message)
			response := fmt.Sprintf("FILE_SERVER_RESPONSE:%s", tcpPort)
			conn.WriteToUDP([]byte(response), clientAddr)
			fmt.Println("Sent discovery response to", clientAddr)
		}
	}
}
