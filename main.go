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

func main() {
	app, err := NewApp()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	err = wails.Run(&options.App{
		Title:     "FoldersGuard",
		Width:     1300,
		Height:    700,
		MinWidth:  700,
		MinHeight: 500,
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
