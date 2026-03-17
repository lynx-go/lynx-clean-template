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

	server.NewKafkaBinderForServer,

	NewComponents,
	NewComponentBuilders,
	NewComponentBuilderSetFunc,
	NewOnStarts,
	NewOnStops,
	NewHealthChecks,
	NewAppConfig,
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
	scheduler *schedule.Scheduler,
	pubSubBroker *pubsub.Broker,
	pubSubBinder *kafka.Binder,
	pubSubRouter *pubsub.Router,
	grpcServer *grpc.Server,
	grpcGatewayServer *server.GRPCGatewayServer,
) []lynx.Component {
	return []lynx.Component{
		scheduler,
		pubSubBroker,
		grpcGatewayServer,
		pubSubRouter,
		pubSubBinder,
		grpcServer,
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

func NewComponentBuilderSetFunc(
	binder *kafka.Binder,
) lynx.ComponentBuilderSetFunc {
	return func() lynx.ComponentBuilderSet {
		return binder.ConsumerBuilders()
	}
}
