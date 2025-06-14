package app

import (
	"context"

	"humpback-agent/api"
	"humpback-agent/interval/controller"
)

type App struct {
	api        *api.Router
	controller *controller.Controller
	stopCh     chan struct{}
}

func InitApp() (*App, error) {
	stopCh := make(chan struct{})
	ctl, err := controller.NewController(stopCh)
	if err != nil {
		return nil, err
	}

	return &App{
		api:        api.InitRouter(ctl),
		controller: ctl,
		stopCh:     stopCh,
	}, nil
}

func (a *App) Startup() error {
	return nil
}

func (a *App) Close(c context.Context) error {
	close(a.stopCh)
	return nil
}
