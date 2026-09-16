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
	"github.com/lynx-go/lynx/contrib/schedule"
	"github.com/lynx-go/lynx/eventbus"
	"github.com/lynx-go/lynx/server/grpc"
)

//go:generate wire

var ProviderSet = wire.NewSet(
	boot.New,
	api.ProviderSet,
	app.ProviderSet,
	infra.ProviderSet,
	domain.ProviderSet,

	NewComponents,
	NewPreStarts,
	NewDrainHooks,
	NewPreStops,
	NewPostStops,
	NewServiceFactories,
	NewAppConfig,
	NewAppBus,
)

func NewAppConfig(app lynx.App) (*config.AppConfig, error) {
	var c config.AppConfig
	if err := config.DecodeLynxConfig(app, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// NewAppBus exposes the application event bus for DI.
func NewAppBus(app lynx.App) eventbus.Bus {
	return app.Bus()
}

func NewComponents(
	scheduler *schedule.Scheduler,
	pubSubRouter *pubsub.Router,
	grpcServer *grpc.Server,
	grpcGatewayServer *server.GRPCGatewayServer,
) []lynx.Service {
	return []lynx.Service{
		scheduler,
		pubSubRouter,
		grpcGatewayServer,
		grpcServer,
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

// NewPostStops 提供进程收尾清理钩子集合；Wire injector 返回的 cleanup
// 由 cmd/server/main.go 直接挂 app.OnPostStop，不经过本集合。
func NewPostStops() boot.PostStopHooks {
	return boot.PostStopHooks{}
}

// NewServiceFactories provides the (empty) service factory set required by boot.New.
func NewServiceFactories() []lynx.ServiceFactory {
	return nil
}
