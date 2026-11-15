package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	udpPort          = 9999
	tcpPort          = 8888
	discoveryMessage = "DISCOVER_FILE_SERVER"
	discoveryReply   = "FILE_SERVER_RESPONSE:"
	messageDelim     = "\r\n"
)

type fileData struct {
	Name     string `json:"name"`
	IsFolder bool   `json:"isFolder"`
	Size     int64  `json:"size"`
}

// listDir returns the visible entries of dir, folders first, then alphabetically.
func listDir(dir string) ([]fileData, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]fileData, 0, len(entries))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		fd := fileData{Name: e.Name(), IsFolder: e.IsDir()}
		if info, err := e.Info(); err == nil && !e.IsDir() {
			fd.Size = info.Size()
		}
		files = append(files, fd)
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsFolder != files[j].IsFolder {
			return files[i].IsFolder
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
	return files, nil
}

var errOutsideRoot = errors.New("path is outside the shared folder")

// resolveInside joins a slash-separated relative path (as sent over the wire) onto root,
// rejecting anything that could escape it, such as "../secret" or "/etc/passwd".
// An empty path refers to root itself.
func resolveInside(root, rel string) (string, error) {
	if rel == "" {
		return root, nil
	}
	local := filepath.FromSlash(rel)
	if !filepath.IsLocal(local) {
		return "", errOutsideRoot
	}
	return filepath.Join(root, local), nil
}

// uniquePath returns a path in dir for name that doesn't exist yet,
// turning "report.pdf" into "report (1).pdf", "report (2).pdf", ...
func uniquePath(dir, name string) string {
	path := filepath.Join(dir, name)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
		path = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
	}
}

var (
	logMu  sync.Mutex
	logDir = func() string {
		if dir, err := os.UserCacheDir(); err == nil {
			return filepath.Join(dir, "transferer", "logs")
		}
		return "logs"
	}()
)

func writeLog(filename, line string) {
	logMu.Lock()
	defer logMu.Unlock()
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Println(err)
		return
	}
	f, err := os.OpenFile(filepath.Join(logDir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	log.New(f, "", log.LstdFlags).Println(line)
}
