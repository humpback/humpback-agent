package app

import (
	"context"

	"humpback-agent/api"
)

type App struct {
	api    *api.Router
	stopCh chan struct{}
}

func InitApp() (*App, error) {
	return nil, nil
}

func (a *App) Startup() error {
	return nil
}

func (a *App) Close(c context.Context) error {
	return nil
}
