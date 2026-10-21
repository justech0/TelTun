// TelTun shares a phone's mobile data with a Windows PC over USB.
// Made by Enes.
package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "1.0.2"

func main() {
	startHidden := false
	for _, a := range os.Args[1:] {
		if a == "--tray" {
			startHidden = true
		}
	}

	app := NewApp()
	app.tray = startTray(app)

	err := wails.Run(&options.App{
		Title:            "TelTun",
		Width:            980,
		Height:           660,
		MinWidth:         860,
		MinHeight:        600,
		Frameless:        true,
		StartHidden:      startHidden,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "teltun-6c1e0a52-enes",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				runtime.WindowShow(app.ctx)
				runtime.WindowUnminimise(app.ctx)
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.Acrylic,
			Theme:                windows.Dark,
		},
		Bind: []interface{}{app},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
