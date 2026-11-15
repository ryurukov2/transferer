package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// config holds user settings that persist between launches.
type config struct {
	mu   sync.RWMutex
	path string

	SharedDir string `json:"sharedDir"`
	SaveDir   string `json:"saveDir"`
}

func loadConfig() *config {
	c := &config{}
	if dir, err := os.UserConfigDir(); err == nil {
		c.path = filepath.Join(dir, "transferer", "config.json")
		if data, err := os.ReadFile(c.path); err == nil {
			if err := json.Unmarshal(data, c); err != nil {
				log.Printf("ignoring invalid config %s: %v", c.path, err)
			}
		}
	}

	home, _ := os.UserHomeDir()
	if c.SharedDir == "" {
		c.SharedDir = home
	}
	if c.SaveDir == "" {
		c.SaveDir = home
		if downloads := filepath.Join(home, "Downloads"); isDir(downloads) {
			c.SaveDir = downloads
		}
	}
	return c
}

func (c *config) sharedDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SharedDir
}

func (c *config) saveDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SaveDir
}

func (c *config) setSharedDir(dir string) error {
	c.mu.Lock()
	c.SharedDir = dir
	c.mu.Unlock()
	return c.save()
}

func (c *config) setSaveDir(dir string) error {
	c.mu.Lock()
	c.SaveDir = dir
	c.mu.Unlock()
	return c.save()
}

func (c *config) save() error {
	if c.path == "" {
		return nil
	}
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
