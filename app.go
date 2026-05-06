package main

import (
	"context"

	"foldersguard/internal/app"
)

type App struct {
	ctx     context.Context
	service app.Service
}

func NewApp() (*App, error) {
	service, err := app.NewService("")
	if err != nil {
		return nil, err
	}
	if err := service.EnsureDataDir(); err != nil {
		return nil, err
	}
	return &App{service: service}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
