package cmd

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/api"
	"github.com/lynx-go/lynx-clean-template/internal/app"
	"github.com/lynx-go/lynx-clean-template/internal/domain"
	"github.com/lynx-go/lynx-clean-template/internal/infra"
	"github.com/lynx-go/lynx-clean-template/internal/infra/server"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/boot"
	"github.com/lynx-go/lynx/contrib/kafka"
)

//go:generate wire

var ProviderSet = wire.NewSet(
	api.ProviderSet,
	app.ProviderSet,
	infra.ProviderSet,
	domain.ProviderSet,
	server.NewKafkaTransportForCLI,
	NewServices,
	NewServiceFactories,
	NewOnStarts,
	NewOnStops,
	NewHealthChecks,
	NewConfiguration,
)

func NewConfiguration(app lynx.App) (*config.AppConfig, error) {
	var c config.AppConfig
	if err := config.UnmarshalConfig(app.Config(), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func NewHealthChecks(app lynx.App) lynx.HealthCheckersFunc {
	return app.HealthCheckers
}

func NewServices(
	broker *pubsub.Broker,
	kafkaT *kafka.Transport,
	router *pubsub.Router,
) []lynx.Service {
	services := []lynx.Service{broker, router}
	if kafkaT != nil {
		services = append(services, kafkaT)
	}
	return services
}

func NewOnStarts() boot.OnStartHooks {
	hooks := boot.OnStartHooks{}
	return hooks
}

func NewOnStops() boot.OnStopHooks {
	hooks := boot.OnStopHooks{}
	return hooks
}

func NewServiceFactories() []lynx.ServiceFactory {
	var factories []lynx.ServiceFactory
	return factories
}
