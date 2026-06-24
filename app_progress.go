package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"foldersguard/internal/progress"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// operationProgressEvent is the Wails event name carrying progress.Event values
// to the WebUI. The frontend subscribes to it and renders the active operation.
const operationProgressEvent = "operation:progress"

type operationHandle struct {
	id     string
	cancel context.CancelFunc
}

var operationCounter atomic.Uint64

func newOperationID() string {
	return fmt.Sprintf("op-%d", operationCounter.Add(1))
}

// beginOperation sets up progress reporting for one long-running operation.
// Operations cannot be cancelled by the user; the context is only cancelled
// internally when the operation finishes (or when the app shuts down) to release
// resources. The returned context carries a progress tracker. The returned
// finish function must be called exactly once with the operation's terminal
// error (nil on success); it emits the terminal progress event and releases the
// operation.
func (a *App) beginOperation(kind string) (context.Context, func(error)) {
	base := a.ctx
	if base == nil {
		base = context.Background()
	}
	ctx, cancel := context.WithCancel(base)
	id := newOperationID()

	emitCtx := a.ctx
	sink := func(event progress.Event) {
		if emitCtx == nil {
			return
		}
		runtime.EventsEmit(emitCtx, operationProgressEvent, event)
	}
	tracker := progress.New(id, kind, sink)
	ctx = progress.NewContext(ctx, tracker)

	a.setCurrentOperation(&operationHandle{id: id, cancel: cancel})
	a.setLongRunningOperationActive(true)
	tracker.Begin()

	finish := func(err error) {
		cancelled := errors.Is(ctx.Err(), context.Canceled)
		tracker.Finish(err, cancelled)
		a.setLongRunningOperationActive(false)
		a.clearCurrentOperation(id)
		cancel()
	}
	return ctx, finish
}

func (a *App) setCurrentOperation(handle *operationHandle) {
	a.currentOperationM.Lock()
	previous := a.currentOperation
	a.currentOperation = handle
	a.currentOperationM.Unlock()
	if previous != nil {
		previous.cancel()
	}
}

func (a *App) clearCurrentOperation(id string) {
	a.currentOperationM.Lock()
	if a.currentOperation != nil && a.currentOperation.id == id {
		a.currentOperation = nil
	}
	a.currentOperationM.Unlock()
}
