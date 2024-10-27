package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	broadcastAddress = "192.168.1.255:9999"
	discoveryMessage = "DISCOVER_FILE_SERVER"
)

type fileData struct {
	Name     string `json:"name"`
	IsFolder bool   `json:"isFolder"`
}

var clientTCPCon net.Conn
var serverAddress string
var receivedFilesDir string

func clientInit() {
	serverAddrs, err := discoverServers()

	if err != nil {
		fmt.Println("Server discovery failed:", err)
		return
	}
	fmt.Println("Discovered servers at:", serverAddrs)

	err = setReceivedFilesDir("received_files")
	if err != nil {
		fmt.Println("Failed to set received files directory. Files will be saved in the executable's directory.", err)
		return
	}
}

func setReceivedFilesDir(newDir string) error {

	stat, err := os.Stat(newDir)
	// 3. file/dir doesn't exist
	if err != nil && os.IsNotExist(err) {
		err := os.Mkdir(newDir, 0777)
		if err != nil {
			return err
		}
		receivedFilesDir = newDir + "/"
		return nil
	}
	// 2. file exists but is not dir
	if err == nil && !stat.IsDir() {
		newDir := newDir + "(1)"
		setReceivedFilesDir(newDir)
	}
	// 4. permission error
	if stat.IsDir() && os.IsPermission(err) {
		newDir := newDir + "(1)"
		setReceivedFilesDir(newDir)
	}
	receivedFilesDir = newDir + "/"
	return nil
}

// discover() returns a string with an IP:PORT format of an active UDP server if one is found
func discover(localIP net.IP) (string, error) {
	localAddr := net.UDPAddr{
		IP:   localIP,
		Port: 0,
	}
	conn, err := net.ListenUDP("udp4", &localAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	broadcastAddr := net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 255),
		Port: 9999,
	}
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.WriteToUDP([]byte(discoveryMessage), &broadcastAddr)
	if err != nil {
		return "", err
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", err
	}

	message := string(buf[:n])
	if strings.HasPrefix(message, "FILE_SERVER_RESPONSE:") {
		tcpPort := strings.TrimPrefix(message, "FILE_SERVER_RESPONSE:")
		serverIP := addr.IP.String()
		return fmt.Sprintf("%s%s", serverIP, tcpPort), nil
	}
	return "", fmt.Errorf("unable to locate server at IP %v", localIP)
}

// getLocalIPs returns a list of local IP addresses of all available IPs for interfaces on a host.
// This can include ethernet, wifi or other IPs that are not loopback and are up.
func getLocalIPs() ([]net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	availableIPs := []net.IP{}

	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // Not an IPv4 address
			}
			availableIPs = append(availableIPs, ip)
		}
	}

	return availableIPs, nil
}

// discoverServers loops through available IP addresses of interfaces on the host and looks for available UDP servers through each IP.
// This is needed because on some computers with more than one interface, inactive interfaces can return 'up' flags similar to active ones.
func discoverServers() ([]string, error) {
	availableServers := []string{}
	localIPs, err := getLocalIPs()
	if err != nil {
		return availableServers, err
	}
	for _, localIP := range localIPs {
		discoveredServer, err := discover(localIP)
		if err != nil {
			fmt.Printf("Error discovering server on %v - %v\n", localIP, err)
			continue
		}
		availableServers = append(availableServers, discoveredServer)
	}
	return availableServers, nil
}

func openTCPConnection(serverAddr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getServerDir() (string, error) {
	if clientTCPCon == nil {
		return "", fmt.Errorf("unable to get existing files - connection is not open, check TCP connection")
	}
	message := "GETDIR" + messageDelim
	_, err := clientTCPCon.Write([]byte(message))
	if err != nil {
		return "", err
	}
	fmt.Println("Sent GETDIR message")
	resp, err := readData(clientTCPCon)
	if err != nil {
		return "", err
	}
	fmt.Println("Received GETDIR response")
	fmt.Println(resp)

	resp = strings.TrimPrefix(resp, "DIR:")

	return resp, nil
}

func getExistingFiles() ([]fileData, error) {
	filesObj := []fileData{}
	if clientTCPCon == nil {
		return filesObj, fmt.Errorf("unable to get existing files - connection is not open, check TCP connection")
	}
	message := "GETFILES" + messageDelim
	_, err := clientTCPCon.Write([]byte(message))
	if err != nil {
		return filesObj, err
	}
	fmt.Println("Sent GETFILES message")
	temp, err := readData(clientTCPCon)
	if err != nil {
		return filesObj, err
	}
	fmt.Println("Received GETFILES response")
	fmt.Println("Files: ", temp)
	filesObj, err = parseFileStr(temp)
	if err != nil {
		return filesObj, err
	}
	return filesObj, nil
}

func checkIfFilePathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func requestFile(filename string) error {
	if clientTCPCon == nil {
		return fmt.Errorf("unable to request file - connection is not open, check TCP connection")
	}
	message := "REQUEST:" + filename + messageDelim
	_, err := clientTCPCon.Write([]byte(message))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(clientTCPCon)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	response = strings.TrimSpace(response)
	fmt.Println(response)
	if strings.HasPrefix(response, "ERROR:") {
		servError := strings.TrimPrefix(response, "ERROR:")
		fmt.Println(servError)
		return fmt.Errorf("%v", servError)
	} else if strings.HasPrefix(response, "SIZE:") {
		sizeStr := strings.TrimPrefix(response, "SIZE:")
		fileSize, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid file size received: %v", err)
		}

		filePath := filepath.Join(receivedFilesDir, filename)
		for checkIfFilePathExists(filePath) {
			filePath = filePath + "(1)"
		}
		fmt.Println("Saving file to:", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		nCopied, err := io.CopyN(file, clientTCPCon, fileSize)
		if err != nil {
			return err
		}
		if nCopied != fileSize {
			return fmt.Errorf("expected to copy %d bytes, but copied %d", fileSize, nCopied)
		}

		fmt.Println("File received successfully")
		return nil
	} else {
		return fmt.Errorf("unexpected response from server: %s", response)
	}
}

func readData(clientTCPCon net.Conn) (string, error) {
	reader := bufio.NewReader(clientTCPCon)
	var result string

	for {
		chunk, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		result += chunk

		if len(chunk) > 0 && chunk[len(chunk)-1] == '\n' {
			break
		}
	}
	return result, nil
}
