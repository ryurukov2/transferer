package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Transferer",
		Width:     1024,
		Height:    700,
		MinWidth:  720,
		MinHeight: 480,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Matches the frontend's background (slate-950) to avoid a flash on startup.
		BackgroundColour: &options.RGBA{R: 2, G: 6, B: 23, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
