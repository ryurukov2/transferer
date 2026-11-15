package main

import (
	"bufio"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// newTestPair connects a fileClient to a fileServer serving dir over an in-memory pipe.
func newTestPair(t *testing.T, dir string) *fileClient {
	t.Helper()
	serverConn, clientConn := net.Pipe()
	srv := &fileServer{dir: func() string { return dir }}
	go srv.handleConn(serverConn)
	c := &fileClient{conn: clientConn, reader: bufio.NewReader(clientConn)}
	t.Cleanup(c.close)
	return c
}

func TestProtocolRoundTrip(t *testing.T) {
	shared, save := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(shared, "a.txt"), []byte("hello"), 0o644)
	os.Mkdir(filepath.Join(shared, "sub"), 0o755)
	os.WriteFile(filepath.Join(save, "a.txt"), []byte("existing"), 0o644)

	c := newTestPair(t, shared)

	dir, err := c.remoteDir()
	if err != nil || dir != shared {
		t.Fatalf("remoteDir = %q, %v", dir, err)
	}

	files, err := c.listFiles("")
	if err != nil {
		t.Fatal(err)
	}
	want := []fileData{{Name: "sub", IsFolder: true}, {Name: "a.txt", Size: 5}}
	if len(files) != 2 || files[0] != want[0] || files[1] != want[1] {
		t.Fatalf("listFiles = %+v", files)
	}

	path, err := c.download("a.txt", save, func(int64, int64) {})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "a (1).txt" {
		t.Errorf("expected collision-safe name, got %s", path)
	}
	if got, _ := os.ReadFile(path); string(got) != "hello" {
		t.Errorf("downloaded content = %q", got)
	}

	// Errors must not desync the connection: later requests still work.
	if _, err := c.download("sub", save, func(int64, int64) {}); err == nil {
		t.Error("downloading a folder should fail")
	}
	if _, err := c.download("../etc/passwd", save, func(int64, int64) {}); err == nil {
		t.Error("path traversal should be rejected")
	}
	if _, err := c.remoteDir(); err != nil {
		t.Errorf("connection unusable after errors: %v", err)
	}
}

func TestSubfolders(t *testing.T) {
	shared, save := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(shared, "sub", "deeper"), 0o755)
	os.WriteFile(filepath.Join(shared, "sub", "deeper", "n.txt"), []byte("nested"), 0o644)

	c := newTestPair(t, shared)

	files, err := c.listFiles("sub")
	if err != nil || len(files) != 1 || files[0].Name != "deeper" {
		t.Fatalf("listFiles(sub) = %+v, %v", files, err)
	}
	if _, err := c.listFiles("../"); err == nil {
		t.Error("listing outside the shared folder should fail")
	}

	path, err := c.download("sub/deeper/n.txt", save, func(int64, int64) {})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(save, "n.txt") {
		t.Errorf("nested file saved to %s, want it directly in the save folder", path)
	}
}

func TestCancelDownload(t *testing.T) {
	shared, save := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(shared, "big.bin"), make([]byte, 4<<20), 0o644)

	c := newTestPair(t, shared)
	_, err := c.download("big.bin", save, func(done, total int64) {
		if done < total {
			c.cancelDownload()
		}
	})
	if !errors.Is(err, errDownloadCancelled) {
		t.Fatalf("err = %v, want errDownloadCancelled", err)
	}
	if entries, _ := os.ReadDir(save); len(entries) != 0 {
		t.Errorf("partial file left behind: %v", entries)
	}
}

func TestServerRejectsTraversal(t *testing.T) {
	c := newTestPair(t, t.TempDir())
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.request("REQUEST:../secret", "SIZE:"); err == nil {
		t.Error("server accepted a path outside the shared folder")
	}
}
