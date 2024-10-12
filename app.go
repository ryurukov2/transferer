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
}

func startServer() {
	wg.Add(1)
	go startUDPServer()
	startTCPServer()
}
func startClient() {
	fmt.Println("client-start")
}
