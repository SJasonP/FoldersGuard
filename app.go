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
	operationGuideGuard         operationGuideCloseGuard
	operationGuideGuardM        sync.RWMutex
}

type operationGuideCloseGuard struct {
	Active bool
	Path   string
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

func (a *App) SetOperationGuideCloseGuardActive(active bool, path string, lang string) {
	a.operationGuideGuardM.Lock()
	defer a.operationGuideGuardM.Unlock()
	a.operationGuideGuard = operationGuideCloseGuard{
		Active: active,
		Path:   path,
		Lang:   lang,
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.longRunningOperationActiveM.RLock()
	active := a.longRunningOperationActive
	a.longRunningOperationActiveM.RUnlock()
	if !active {
		a.operationGuideGuardM.RLock()
		guideGuard := a.operationGuideGuard
		a.operationGuideGuardM.RUnlock()
		if !guideGuard.Active {
			return false
		}
		return a.confirmCloseWithOperationGuide(ctx, guideGuard)
	}
	_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   "FoldersGuard is working",
		Message: "An operation is still running. FoldersGuard will block normal close until it finishes.",
		Buttons: []string{"OK"},
	})
	return true
}

func (a *App) confirmCloseWithOperationGuide(ctx context.Context, guard operationGuideCloseGuard) bool {
	title := "Operation guide still needs attention"
	message := "An operation guide was generated for manual encrypted-content changes. Make sure you have saved or followed it before closing FoldersGuard."
	closeButton := "Close"
	stayButton := "Stay"
	if guard.Lang == app.LanguageZHCN {
		title = "操作指南仍需处理"
		message = "本次更改已生成用于手动处理加密内容的操作指南。关闭 FoldersGuard 前，请确认你已经保存或完成指南中的步骤。"
		closeButton = "关闭"
		stayButton = "留在应用内"
	}
	if guard.Path != "" {
		message += "\n\n" + guard.Path
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
