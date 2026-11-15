package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// fileServer answers discovery broadcasts over UDP and serves the shared folder over TCP.
//
// TCP protocol (one request per line, terminated by "\r\n"). Paths are slash-separated
// and relative to the shared folder; an empty path means the shared folder itself.
//
//	GETDIR            -> "DIR:<absolute shared folder>\n"
//	GETFILES:<path>   -> "FILES:<json []fileData>\n"
//	REQUEST:<path>    -> "SIZE:<bytes>\n" followed by the raw file contents
//
// Any request may instead be answered with "ERROR:<message>\n".
type fileServer struct {
	dir func() string
	// onActivity, if set, is called when a client connects, disconnects or downloads a file.
	onActivity func(serverActivity)

	mu  sync.Mutex
	udp *net.UDPConn
	tcp net.Listener
}

func (s *fileServer) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tcp != nil {
		return nil
	}

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
	if err != nil {
		return fmt.Errorf("port %d is unavailable (is another server running?): %w", tcpPort, err)
	}
	udp, err := net.ListenUDP("udp4", &net.UDPAddr{Port: udpPort})
	if err != nil {
		tcp.Close()
		return fmt.Errorf("discovery port %d is unavailable: %w", udpPort, err)
	}

	s.tcp, s.udp = tcp, udp
	go s.serveDiscovery(udp)
	go s.serveTCP(tcp)
	log.Printf("server listening on tcp :%d, udp :%d", tcpPort, udpPort)
	return nil
}

func (s *fileServer) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tcp != nil {
		s.tcp.Close()
		s.tcp = nil
	}
	if s.udp != nil {
		s.udp.Close()
		s.udp = nil
	}
}

func (s *fileServer) serveDiscovery(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return // connection closed by stop()
		}
		if string(buf[:n]) == discoveryMessage {
			reply := fmt.Sprintf("%s:%d", discoveryReply, tcpPort)
			conn.WriteToUDP([]byte(reply), addr)
		}
	}
}

func (s *fileServer) serveTCP(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return // listener closed by stop()
		}
		go s.handleConn(conn)
	}
}

// serverActivity describes something a client did, for display in the UI.
type serverActivity struct {
	Kind   string `json:"kind"` // "connected", "disconnected" or "download"
	Client string `json:"client"`
	Path   string `json:"path,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

func (s *fileServer) report(a serverActivity) {
	if s.onActivity != nil {
		s.onActivity(a)
	}
}

func (s *fileServer) handleConn(conn net.Conn) {
	client, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	s.report(serverActivity{Kind: "connected", Client: client})
	defer s.report(serverActivity{Kind: "disconnected", Client: client})
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		msg := strings.TrimRight(line, messageDelim)
		writeLog("requests.log", fmt.Sprintf("%s: %s", client, msg))

		cmd, arg, _ := strings.Cut(msg, ":")
		switch cmd {
		case "GETDIR":
			fmt.Fprintf(conn, "DIR:%s\n", s.dir())
		case "GETFILES":
			s.sendFileList(conn, arg)
		case "REQUEST":
			if size, ok := s.sendFile(conn, arg); ok {
				s.report(serverActivity{Kind: "download", Client: client, Path: arg, Size: size})
			}
		default:
			sendError(conn, "unknown request")
		}
	}
}

func (s *fileServer) sendFileList(conn net.Conn, rel string) {
	dir, err := resolveInside(s.dir(), rel)
	if err != nil {
		sendError(conn, err.Error())
		return
	}
	files, err := listDir(dir)
	if err != nil {
		sendError(conn, "cannot read folder")
		return
	}
	data, err := json.Marshal(files)
	if err != nil {
		sendError(conn, "cannot encode file list")
		return
	}
	fmt.Fprintf(conn, "FILES:%s\n", data)
}

// sendFile streams the file at rel and reports whether the whole file was sent.
func (s *fileServer) sendFile(conn net.Conn, rel string) (int64, bool) {
	path, err := resolveInside(s.dir(), rel)
	if err != nil || rel == "" {
		sendError(conn, "invalid file path")
		return 0, false
	}
	file, err := os.Open(path)
	if err != nil {
		sendError(conn, "file not found")
		return 0, false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		sendError(conn, "cannot read file")
		return 0, false
	}
	if info.IsDir() {
		sendError(conn, "folders can't be downloaded yet")
		return 0, false
	}

	fmt.Fprintf(conn, "SIZE:%d\n", info.Size())
	if _, err := io.Copy(conn, file); err != nil {
		log.Printf("sending %s failed: %v", rel, err)
		return 0, false
	}
	return info.Size(), true
}

func sendError(conn net.Conn, msg string) {
	fmt.Fprintf(conn, "ERROR:%s\n", msg)
}
