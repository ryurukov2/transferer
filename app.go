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
	wg.Wait()
	fmt.Println("asd")
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
func (a *App) SetClientConnection(connAddress string) {
	serverAddress = connAddress
	conn, err := openTCPConnection(serverAddress)
	if err != nil {
		fmt.Println(err)
	}
	wg.Add(1)
	clientTCPCon = conn
}
func (a *App) GetFiles() []fileData {
	files, err := getExistingFiles()
	if err != nil {
		fmt.Println(err)
		return []fileData{}
	}
	return files

}

func startServer() {
	wg.Add(1)
	go startUDPServer()
	startTCPServer()
}
func startClient() {
	fmt.Println("client-start")
	clientInit()
}

func stopServers() {
	fmt.Println("Stopserv")

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
}
