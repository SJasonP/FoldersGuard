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
	uiLang                      string
	uiLangM                     sync.RWMutex
	manualContentGuideGuard     manualContentGuideCloseGuard
	manualContentGuideGuardM    sync.RWMutex
	currentOperation            *operationHandle
	currentOperationM           sync.Mutex
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
	a.restoreWindowPlacement(ctx)
}

func (a *App) setLongRunningOperationActive(active bool) {
	a.longRunningOperationActiveM.Lock()
	defer a.longRunningOperationActiveM.Unlock()
	a.longRunningOperationActive = active
}

// SetUILanguage records the current WebUI language so native dialogs (such as
// the close guard shown while an operation is running) can be localized.
func (a *App) SetUILanguage(lang string) {
	a.uiLangM.Lock()
	defer a.uiLangM.Unlock()
	a.uiLang = lang
}

func (a *App) currentUILanguage() string {
	a.uiLangM.RLock()
	defer a.uiLangM.RUnlock()
	return a.uiLang
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
			a.saveWindowPlacement(ctx)
			return false
		}
		preventClose := a.confirmCloseWithManualContentGuide(ctx, guideGuard)
		if !preventClose {
			a.saveWindowPlacement(ctx)
		}
		return preventClose
	}
	a.warnLongRunningOperationClose(ctx)
	return true
}

// warnLongRunningOperationClose tells the user that closing is blocked while an
// operation runs, and that forcing the app to quit anyway is entirely at their
// own risk. Operations cannot be cancelled, so the only safe path is to wait.
func (a *App) warnLongRunningOperationClose(ctx context.Context) {
	title := "FoldersGuard is working"
	message := "An operation is still running and cannot be cancelled. FoldersGuard blocks closing the window and quitting the app until it finishes.\n\n" +
		"If you force the app to quit anyway — for example, Force Quit or killing the process — the encrypted output and your source files may be left incomplete or damaged. Any errors or data loss that result are entirely your own responsibility and are not the responsibility of FoldersGuard or its developers."
	okButton := "OK"
	if a.currentUILanguage() == app.LanguageZHCN {
		title = "FoldersGuard 正在执行操作"
		message = "操作仍在进行中，且无法取消。在操作完成前，FoldersGuard 会阻止关闭窗口和退出应用。\n\n" +
			"如果你仍然强行退出（例如“强制退出”或直接结束进程），加密输出和你的源文件可能会处于不完整或损坏的状态。由此产生的任何错误或数据丢失将完全由你自行承担，与 FoldersGuard 及其开发者无关。"
		okButton = "我已知晓"
	}
	_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   title,
		Message: message,
		Buttons: []string{okButton},
	})
}

func (a *App) initialWindowSize() (int, int) {
	placement, ok, err := a.service.ReadWindowPlacement(minWindowWidth, minWindowHeight)
	if err != nil || !ok {
		return defaultWindowWidth, defaultWindowHeight
	}
	return placement.Width, placement.Height
}

func (a *App) restoreWindowPlacement(ctx context.Context) {
	placement, ok, err := a.service.ReadWindowPlacement(minWindowWidth, minWindowHeight)
	if err != nil || !ok {
		return
	}
	runtime.WindowSetSize(ctx, placement.Width, placement.Height)
	runtime.WindowSetPosition(ctx, placement.X, placement.Y)
	if placement.Maximised {
		runtime.WindowMaximise(ctx)
	}
}

func (a *App) saveWindowPlacement(ctx context.Context) {
	if runtime.WindowIsFullscreen(ctx) || runtime.WindowIsMinimised(ctx) {
		return
	}
	width, height := runtime.WindowGetSize(ctx)
	x, y := runtime.WindowGetPosition(ctx)
	placement := app.WindowPlacement{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Maximised: runtime.WindowIsMaximised(ctx),
	}
	_ = a.service.SaveWindowPlacement(placement, minWindowWidth, minWindowHeight)
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
