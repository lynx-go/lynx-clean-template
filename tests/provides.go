package tests

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/api"
	"github.com/lynx-go/lynx-clean-template/internal/app"
	"github.com/lynx-go/lynx-clean-template/internal/domain"
	"github.com/lynx-go/lynx-clean-template/internal/infra"
	"github.com/lynx-go/lynx-clean-template/internal/infra/server"
	config "github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/contrib/kafka"
)

var ProviderSet = wire.NewSet(
	api.ProviderSet,
	app.ProviderSet,
	infra.ProviderSet,
	domain.ProviderSet,
	server.NewKafkaBinderForCLI,
	NewComponents,
	NewComponentBuilders,
	NewComponentBuilderSetFunc,
	NewOnStarts,
	NewOnStops,
	NewHealthChecks,
	NewAppConfig,
	NewTestingSuite,
)

func NewAppConfig(app lynx.Lynx) (*config.AppConfig, error) {
	var c config.AppConfig
	if err := app.Config().Unmarshal(&c, lynx.TagNameJSON); err != nil {
		return nil, err
	}
	return &c, nil
}

func NewHealthChecks(app lynx.Lynx) lynx.HealthCheckFunc {
	return app.HealthCheckFunc()
}

func NewComponents(
	broker *pubsub.Broker,
	binder *kafka.Binder,
	router *pubsub.Router,
) []lynx.Component {
	return []lynx.Component{
		broker,
		binder,
		router,
	}
}

func NewOnStarts() lynx.OnStartHooks {
	hooks := lynx.OnStartHooks{}
	return hooks
}

func NewOnStops() lynx.OnStopHooks {
	hooks := lynx.OnStopHooks{}
	return hooks
}

func NewComponentBuilders() []lynx.ComponentBuilder {
	var builders []lynx.ComponentBuilder
	return builders
}

func NewComponentBuilderSetFunc() lynx.ComponentBuilderSetFunc {
	return func() lynx.ComponentBuilderSet {
		var builders []lynx.ComponentBuilder
		return builders
	}
}
