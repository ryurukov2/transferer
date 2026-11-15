package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is bound to the frontend; its exported methods are callable from JavaScript.
// Methods that return an error reject the JS promise with the error message.
type App struct {
	ctx    context.Context
	cfg    *config
	server *fileServer
	client *fileClient
}

func NewApp() *App {
	cfg := loadConfig()
	return &App{
		cfg:    cfg,
		server: &fileServer{dir: cfg.sharedDir},
		client: &fileClient{},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Emitted as "server-activity" events: {kind, client, path, size}.
	a.server.onActivity = func(act serverActivity) {
		runtime.EventsEmit(ctx, "server-activity", act)
	}
}

func (a *App) shutdown(ctx context.Context) { a.Stop() }

// Stop shuts down the server and closes any client connection.
func (a *App) Stop() {
	a.server.stop()
	a.client.close()
}

// ---- Server mode ----

func (a *App) StartServer() error { return a.server.start() }

// GetLocalAddresses returns the addresses other machines can use to reach this one.
func (a *App) GetLocalAddresses() []string {
	networks, err := localNetworks()
	if err != nil {
		return nil
	}
	addrs := make([]string, len(networks))
	for i, n := range networks {
		addrs[i] = fmt.Sprintf("%s:%d", n.IP, tcpPort)
	}
	return addrs
}

func (a *App) GetSharedDir() string { return a.cfg.sharedDir() }

// ListSharedFiles lists a folder inside the shared folder ("" for the shared folder itself).
func (a *App) ListSharedFiles(rel string) ([]fileData, error) {
	dir, err := resolveInside(a.cfg.sharedDir(), rel)
	if err != nil {
		return nil, err
	}
	return listDir(dir)
}

// OpenSharedDir opens the shared folder in the system file manager.
func (a *App) OpenSharedDir() error { return openInFileManager(a.cfg.sharedDir()) }

// SelectSharedDir opens a folder picker; it returns "" if the user cancels.
func (a *App) SelectSharedDir() (string, error) {
	dir, err := a.pickFolder("Select a folder to share", a.cfg.sharedDir())
	if err != nil || dir == "" {
		return "", err
	}
	return dir, a.cfg.setSharedDir(dir)
}

// ---- Client mode ----

func (a *App) DiscoverServers() ([]string, error) { return discoverServers(2 * time.Second) }

func (a *App) Connect(addr string) error { return a.client.connect(addr) }

func (a *App) GetRemoteDir() (string, error) { return a.client.remoteDir() }

// ListRemoteFiles lists a folder on the server, relative to its shared folder.
func (a *App) ListRemoteFiles(rel string) ([]fileData, error) { return a.client.listFiles(rel) }

// DownloadFile saves the remote file at rel into the save folder and returns its local path.
// Progress is emitted as "download-progress" events: {path, done, total}.
func (a *App) DownloadFile(rel string) (string, error) {
	return a.client.download(rel, a.cfg.saveDir(), func(done, total int64) {
		runtime.EventsEmit(a.ctx, "download-progress", map[string]any{
			"path": rel, "done": done, "total": total,
		})
	})
}

// CancelDownload aborts the download in progress, if any.
func (a *App) CancelDownload() { a.client.cancelDownload() }

func (a *App) GetSaveDir() string { return a.cfg.saveDir() }

// SelectSaveDir opens a folder picker; it returns "" if the user cancels.
func (a *App) SelectSaveDir() (string, error) {
	dir, err := a.pickFolder("Select where to save downloads", a.cfg.saveDir())
	if err != nil || dir == "" {
		return "", err
	}
	return dir, a.cfg.setSaveDir(dir)
}

// OpenSaveDir opens the save folder in the system file manager.
func (a *App) OpenSaveDir() error { return openInFileManager(a.cfg.saveDir()) }

func (a *App) pickFolder(title, current string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: current,
	})
}
