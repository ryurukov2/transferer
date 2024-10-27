package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.EventsOn(ctx, "start-server", func(optionalData ...interface{}) {
		startServer()
	})
	runtime.EventsOn(ctx, "start-client", func(optionalData ...interface{}) {
		startClient()
	})
	runtime.EventsOn(ctx, "stop-servers", func(optionalData ...interface{}) {
		stopServers()
	})
}

func (a *App) shutdown(ctx context.Context) {
	stopServers()
}

func (a *App) ReqFile(filename string) string {
	err := requestFile(filename)
	if err != nil {
		return fmt.Sprintf("Error requesting the file, %v", err)
	}
	return "File downloaded successfully."
}

func (a *App) DiscServers() []string {

	servers, err := discoverServers()
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return servers
}
func (a *App) SetClientConnection(connAddress string) bool {
	serverAddress = connAddress
	conn, err := openTCPConnection(serverAddress)
	if err != nil {
		fmt.Printf("Error starting client connection, %v\n", err)
		return false
	}
	clientTCPCon = conn
	return true
}

func (a *App) GetServerDirectory() string {
	servDir, err := getServerDir()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return servDir
}

func (a *App) GetFilesFromClient() []fileData {
	files, err := getExistingFiles()
	if err != nil {
		fmt.Println(err)
		return []fileData{}
	}
	return files
}

func (a *App) GetFilesFromServer(dir string) []fileData {
	filesObj := []fileData{}
	fileStr, err := getFiles(dir)
	if err != nil {
		fmt.Printf("error getting files - %v\n", err)
		return filesObj
	}
	filesObj, err = parseFileStr(fileStr)
	if err != nil {
		fmt.Printf("error parsing files string to file data - %v\n", err)
		return filesObj
	}
	return filesObj

}

func (a *App) GetCurrentServerDir() string {
	return serverDir
}
func (a *App) SelectFolder() (string, error) {
	folderPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a folder to share",
	})
	if err != nil {
		return "", err
	}
	setServerDir(folderPath)
	return folderPath, nil
}

func startServer() {
	serverInit()
}
func startClient() {
	clientInit()
}

func stopServers() {
	if tcpListener != nil {
		fmt.Println("Stopping TCP server...")
		tcpListener.Close()
		tcpListener = nil
	}
	if udpConn != nil {
		fmt.Println("Stopping UDP server...")
		udpConn.Close()
		udpConn = nil
	}
	if clientTCPCon != nil {
		fmt.Println("Closing client TCP connection...")
		clientTCPCon.Close()
		clientTCPCon = nil
	}
	if serverAddress != "" {
		serverAddress = ""
	}
}
