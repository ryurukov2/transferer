package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	udpPort = ":9999"
	tcpPort = ":8888"
)

var udpConn *net.UDPConn
var tcpListener net.Listener
var serverDir string = "."

func serverInit() {
	go startUDPServer()
	startTCPServer()
	err := setLogDir()
	if err != nil {
		fmt.Println("Error setting log directory, logs will not be saved - ", err)
	}
}

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
	udpConn = conn
	defer conn.Close()

	fmt.Println("UDP server listening on", udpPort)

	buf := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			conn.Close()
			return
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
	tcpListener = listener
	defer listener.Close()

	fmt.Println("TCP server listening on", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			return
		}
		go handleTCPConnection(conn)
	}
}

// Handles TCP connection for sending a file.
func handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	var buffer string
	for {

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from TCP connection:", err)
			conn.Close()
			return
		}
		buffer += string(buf[:n])
		go writeLog("requestMessages.log", buffer)
		fmt.Println(buffer)
		for {
			idx := strings.Index(buffer, "\r\n")
			fmt.Println(idx)
			if idx == -1 {
				break
			}

			message := buffer[:idx]
			buffer = buffer[idx+2:]

			fmt.Println("Received message:", message)
			handleTCPRequest(conn, message)
		}
	}
}

func handleTCPRequest(conn net.Conn, message string) {
	if strings.HasPrefix(message, "GETDIR") {
		sendFileDir(conn)
	} else if strings.HasPrefix(message, "GETFILES") {
		sendExistingFiles(conn)
	} else if strings.HasPrefix(message, "REQUEST:") {
		filename := strings.TrimPrefix(message, "REQUEST:")
		sendFile(conn, filename)
	}
}
func sendFileDir(conn net.Conn) {
	fmt.Println("Sending current directory - ")
	dirStr := fmt.Sprintf("DIR:%v\n", serverDir)
	conn.Write([]byte(dirStr))
}
func sendFile(conn net.Conn, filename string) {
	fmt.Println("Client requested file:", filename)
	filePath := filepath.Join(serverDir, "/", filename)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		conn.Write([]byte("ERROR: File not found"))
		return
	}
	defer file.Close()
	fStat, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file data:", err)
		conn.Write([]byte("ERROR: Error getting file data"))
		return
	}
	fileSize := strconv.FormatInt(fStat.Size(), 10)
	conn.Write([]byte("SIZE:" + fileSize + "\n"))
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file:", err)
		return
	}
	fmt.Println("File sent successfully")
}

func sendExistingFiles(conn net.Conn) {
	fileStr, err := getFiles(serverDir)
	if err != nil {
		fmt.Printf("unable to access directory - %v\n", err)
		return
	}
	_, err = conn.Write([]byte(fileStr))
	if err != nil {
		fmt.Printf("unable to send list of files via the TCP connection - %v", err)
	}
}

// Returns a string with all the files in a specified directory, deliminated by two commas
// (this is because commas are disallowed characters for file names in both windows and linux)
func getFiles(dir string) (string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return "", err
	}
	fileStr, err := readFilesToStr(f)
	if err != nil {
		return "", err
	}
	fileStr += "\n"
	return fileStr, nil
}

func setServerDir(dir string) {
	serverDir = dir
}
