package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

const (
	defaultWindowWidth  = 1300
	defaultWindowHeight = 700
	minWindowWidth      = 700
	minWindowHeight     = 500
)

func main() {
	app, err := NewApp()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	windowWidth, windowHeight := app.initialWindowSize()
	err = wails.Run(&options.App{
		Title:     "FoldersGuard",
		Width:     windowWidth,
		Height:    windowHeight,
		MinWidth:  minWindowWidth,
		MinHeight: minWindowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:     app.startup,
		OnBeforeClose: app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
