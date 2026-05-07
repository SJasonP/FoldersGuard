package main

import (
	"context"
	"sync"

	"foldersguard/internal/app"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                         context.Context
	service                     app.Service
	startupError                error
	longRunningOperationActive  bool
	longRunningOperationActiveM sync.RWMutex
}

func NewApp() (*App, error) {
	service, err := app.NewService("")
	if err != nil {
		return nil, err
	}
	fgApp := &App{service: service}
	if err := service.EnsureDataDir(); err != nil {
		fgApp.startupError = err
	}
	return fgApp, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SetLongRunningOperationActive(active bool) {
	a.longRunningOperationActiveM.Lock()
	defer a.longRunningOperationActiveM.Unlock()
	a.longRunningOperationActive = active
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.longRunningOperationActiveM.RLock()
	active := a.longRunningOperationActive
	a.longRunningOperationActiveM.RUnlock()
	if !active {
		return false
	}
	_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   "FoldersGuard is working",
		Message: "An operation is still running. FoldersGuard will block normal close until it finishes.",
		Buttons: []string{"OK"},
	})
	return true
}
