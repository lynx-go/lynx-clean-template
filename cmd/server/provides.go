package main

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
	"github.com/lynx-go/lynx/boot"
	"github.com/lynx-go/lynx/contrib/kafka"
	"github.com/lynx-go/lynx/contrib/schedule"
	"github.com/lynx-go/lynx/server/grpc"
)

//go:generate wire

var ProviderSet = wire.NewSet(
	boot.New,
	api.ProviderSet,
	app.ProviderSet,
	infra.ProviderSet,
	domain.ProviderSet,

	server.NewKafkaTransportForServer,

	NewServices,
	NewServiceFactories,
	NewOnStarts,
	NewOnStops,
	NewHealthChecks,
	NewAppConfig,
)

func NewAppConfig(app lynx.App) (*config.AppConfig, error) {
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
	scheduler *schedule.Scheduler,
	pubSubBroker *pubsub.Broker,
	pubSubTransport *kafka.Transport,
	pubSubRouter *pubsub.Router,
	grpcServer *grpc.Server,
	grpcGatewayServer *server.GRPCGatewayServer,
) []lynx.Service {
	services := []lynx.Service{
		scheduler,
		pubSubBroker,
		grpcGatewayServer,
		pubSubRouter,
	}
	if pubSubTransport != nil {
		services = append(services, pubSubTransport)
	}
	services = append(services, grpcServer)
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
