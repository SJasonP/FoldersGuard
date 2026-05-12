package main

import (
	"context"
	"sync"

	"foldersguard/internal/app"
	fgdb "foldersguard/internal/db"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                         context.Context
	service                     app.Service
	startupError                error
	longRunningOperationActive  bool
	longRunningOperationActiveM sync.RWMutex
	manualContentGuideGuard     manualContentGuideCloseGuard
	manualContentGuideGuardM    sync.RWMutex
}

type manualContentGuideCloseGuard struct {
	Active bool
	Lang   string
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
	if err := fgdb.SQLCipherAvailable(); err != nil && fgApp.startupError == nil {
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

func (a *App) SetManualContentGuideCloseGuardActive(active bool, lang string) {
	a.manualContentGuideGuardM.Lock()
	defer a.manualContentGuideGuardM.Unlock()
	a.manualContentGuideGuard = manualContentGuideCloseGuard{
		Active: active,
		Lang:   lang,
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.longRunningOperationActiveM.RLock()
	active := a.longRunningOperationActive
	a.longRunningOperationActiveM.RUnlock()
	if !active {
		a.manualContentGuideGuardM.RLock()
		guideGuard := a.manualContentGuideGuard
		a.manualContentGuideGuardM.RUnlock()
		if !guideGuard.Active {
			return false
		}
		return a.confirmCloseWithManualContentGuide(ctx, guideGuard)
	}
	_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   "FoldersGuard is working",
		Message: "An operation is still running. FoldersGuard will block normal close until it finishes.",
		Buttons: []string{"OK"},
	})
	return true
}

func (a *App) confirmCloseWithManualContentGuide(ctx context.Context, guard manualContentGuideCloseGuard) bool {
	title := "Manual content guide still needs attention"
	message := "Manual encrypted-content instructions are still open in FoldersGuard. Make sure you have finished them or kept the window available before closing."
	closeButton := "Close"
	stayButton := "Stay"
	if guard.Lang == app.LanguageZHCN {
		title = "手动处理指南仍需处理"
		message = "用于手动处理加密内容的指南仍在 FoldersGuard 内显示。关闭前，请确认你已经完成指南中的步骤，或仍能查看当前窗口。"
		closeButton = "关闭"
		stayButton = "留在应用内"
	}
	button, _ := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{closeButton, stayButton},
		DefaultButton: stayButton,
		CancelButton:  stayButton,
	})
	return button != closeButton
}
