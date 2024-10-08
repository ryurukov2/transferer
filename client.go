package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	broadcastAddress = "192.168.1.255:9999"
	discoveryMessage = "DISCOVER_FILE_SERVER"
)

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
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
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

func requestFile(serverAddr, filename string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(filename))
	if err != nil {
		return err
	}

	file, err := os.Create("received_" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, conn)
	if err != nil {
		os.Remove("received_" + filename)
		return err
	}

	fmt.Println("File received successfully")
	return nil
}
