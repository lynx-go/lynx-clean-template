package tests

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/api"
	"github.com/lynx-go/lynx-clean-template/internal/app"
	"github.com/lynx-go/lynx-clean-template/internal/domain"
	"github.com/lynx-go/lynx-clean-template/internal/infra"
	config "github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/boot"
	"github.com/lynx-go/lynx/eventbus"
)

var ProviderSet = wire.NewSet(
	api.ProviderSet,
	app.ProviderSet,
	infra.ProviderSet,
	domain.ProviderSet,
	NewComponents,
	NewPreStarts,
	NewDrainHooks,
	NewPreStops,
	NewPostStops,
	NewAppConfig,
	NewTestingSuite,
	NewAppBus,
)

func NewAppConfig(app lynx.App) (*config.AppConfig, error) {
	var c config.AppConfig
	if err := config.DecodeLynxConfig(app, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func NewAppBus(app lynx.App) eventbus.Bus {
	return app.Bus()
}

func NewComponents(
	router *pubsub.Router,
) []lynx.Service {
	return []lynx.Service{
		router,
	}
}

func NewPreStarts() boot.PreStartHooks {
	return boot.PreStartHooks{}
}

func NewDrainHooks() boot.DrainHooks {
	return boot.DrainHooks{}
}

func NewPreStops() boot.PreStopHooks {
	return boot.PreStopHooks{}
}

func NewPostStops() boot.PostStopHooks {
	return boot.PostStopHooks{}
}
