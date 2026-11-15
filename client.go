package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const requestTimeout = 10 * time.Second

var (
	errNotConnected      = errors.New("not connected to a server")
	errDownloadCancelled = errors.New("download cancelled")
)

// fileClient holds a single TCP connection to a fileServer.
// All requests are serialised by mu, and a single buffered reader is kept for the
// lifetime of the connection so no response bytes are lost between requests.
type fileClient struct {
	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader

	// cancelMu guards the fields used to abort a download from another goroutine.
	cancelMu    sync.Mutex
	downloading net.Conn
	cancelled   bool
}

func (c *fileClient) connect(addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to %s: %w", addr, err)
	}
	c.conn, c.reader = conn, bufio.NewReader(conn)
	return nil
}

func (c *fileClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
}

func (c *fileClient) closeLocked() {
	if c.conn != nil {
		c.conn.Close()
		c.conn, c.reader = nil, nil
	}
}

// request sends cmd and returns the reply line with wantPrefix removed. Callers must hold mu.
func (c *fileClient) request(cmd, wantPrefix string) (string, error) {
	if c.conn == nil {
		return "", errNotConnected
	}
	c.conn.SetDeadline(time.Now().Add(requestTimeout))
	defer c.conn.SetDeadline(time.Time{})

	if _, err := io.WriteString(c.conn, cmd+messageDelim); err != nil {
		c.closeLocked()
		return "", fmt.Errorf("connection lost: %w", err)
	}
	line, err := c.reader.ReadString('\n')
	if err != nil {
		c.closeLocked()
		return "", fmt.Errorf("connection lost: %w", err)
	}
	line = strings.TrimRight(line, messageDelim)

	if msg, ok := strings.CutPrefix(line, "ERROR:"); ok {
		return "", errors.New(msg)
	}
	payload, ok := strings.CutPrefix(line, wantPrefix)
	if !ok {
		return "", fmt.Errorf("unexpected response from server: %q", line)
	}
	return payload, nil
}

func (c *fileClient) remoteDir() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.request("GETDIR", "DIR:")
}

// listFiles lists the folder at rel (slash-separated, relative to the shared folder).
func (c *fileClient) listFiles(rel string) ([]fileData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	payload, err := c.request("GETFILES:"+rel, "FILES:")
	if err != nil {
		return nil, err
	}
	var files []fileData
	if err := json.Unmarshal([]byte(payload), &files); err != nil {
		return nil, fmt.Errorf("invalid file list from server: %w", err)
	}
	return files, nil
}

// download fetches the remote file at rel into saveDir and returns the local path.
// progress is called periodically with the number of bytes received so far.
func (c *fileClient) download(rel, saveDir string, progress func(done, total int64)) (string, error) {
	if rel == "" || !filepath.IsLocal(filepath.FromSlash(rel)) {
		return "", fmt.Errorf("invalid file path %q", rel)
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancelMu.Lock()
	c.cancelled = false
	c.downloading = c.conn
	c.cancelMu.Unlock()
	defer func() {
		c.cancelMu.Lock()
		c.downloading = nil
		c.cancelMu.Unlock()
	}()

	sizeStr, err := c.request("REQUEST:"+rel, "SIZE:")
	if err != nil {
		return "", err
	}
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid file size from server: %q", sizeStr)
	}

	dest := uniquePath(saveDir, path.Base(rel))
	file, err := os.Create(dest)
	if err != nil {
		// The server is already streaming; drop the connection so it doesn't desync.
		c.closeLocked()
		return "", err
	}
	defer file.Close()

	pw := &progressWriter{w: file, total: size, report: progress}
	if _, err := io.CopyN(pw, c.reader, size); err != nil {
		file.Close()
		os.Remove(dest)
		c.closeLocked()
		if c.wasCancelled() {
			return "", errDownloadCancelled
		}
		return "", fmt.Errorf("transfer interrupted: %w", err)
	}
	progress(size, size)
	return dest, nil
}

// cancelDownload aborts an in-progress download by closing its connection.
// It deliberately doesn't take mu, which the download is holding.
func (c *fileClient) cancelDownload() {
	c.cancelMu.Lock()
	defer c.cancelMu.Unlock()
	if c.downloading != nil {
		c.cancelled = true
		c.downloading.Close()
	}
}

func (c *fileClient) wasCancelled() bool {
	c.cancelMu.Lock()
	defer c.cancelMu.Unlock()
	return c.cancelled
}

// progressWriter forwards writes and reports progress at most every 100ms.
type progressWriter struct {
	w           io.Writer
	done, total int64
	last        time.Time
	report      func(done, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.done += int64(n)
	if now := time.Now(); now.Sub(p.last) > 100*time.Millisecond {
		p.last = now
		p.report(p.done, p.total)
	}
	return n, err
}

// discoverServers broadcasts on every active IPv4 network in parallel and
// returns the sorted, de-duplicated "ip:port" addresses of servers that replied.
func discoverServers(timeout time.Duration) ([]string, error) {
	networks, err := localNetworks()
	if err != nil {
		return nil, err
	}

	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		found = map[string]bool{}
	)
	for _, n := range networks {
		wg.Add(1)
		go func(n *net.IPNet) {
			defer wg.Done()
			for _, addr := range discoverOn(n, timeout) {
				mu.Lock()
				found[addr] = true
				mu.Unlock()
			}
		}(n)
	}
	wg.Wait()

	servers := make([]string, 0, len(found))
	for addr := range found {
		servers = append(servers, addr)
	}
	sort.Strings(servers)
	return servers, nil
}

// discoverOn sends a discovery broadcast on network n and collects every reply until timeout.
func discoverOn(n *net.IPNet, timeout time.Duration) []string {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: n.IP})
	if err != nil {
		return nil
	}
	defer conn.Close()

	if _, err := conn.WriteToUDP([]byte(discoveryMessage), &net.UDPAddr{IP: broadcastAddr(n), Port: udpPort}); err != nil {
		return nil
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	var servers []string
	buf := make([]byte, 1024)
	for {
		size, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return servers // deadline reached
		}
		if port, ok := strings.CutPrefix(string(buf[:size]), discoveryReply); ok {
			servers = append(servers, addr.IP.String()+port)
		}
	}
}

// localNetworks returns the IPv4 networks of all interfaces that are up and not loopback.
func localNetworks() ([]*net.IPNet, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var networks []*net.IPNet
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
				networks = append(networks, &net.IPNet{IP: ipNet.IP.To4(), Mask: ipNet.Mask[len(ipNet.Mask)-4:]})
			}
		}
	}
	return networks, nil
}

func broadcastAddr(n *net.IPNet) net.IP {
	ip := make(net.IP, 4)
	for i := range ip {
		ip[i] = n.IP[i] | ^n.Mask[i]
	}
	return ip
}
