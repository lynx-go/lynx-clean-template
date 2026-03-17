package infra

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/internal/infra/server"
)

var ProviderSet = wire.NewSet(

	server.NewGRPCGatewayServer,
	server.NewGRPCServer,
	server.NewAuthValidator,
	server.NewScheduler,
	clients.NewDataClients,

	server.NewPubSub,
	server.NewPublisher,
	server.NewPubSubRouter,

	NewDomainLogger,
	NewDomainEventPublisher,
	NewPasswordHasher,
	NewEmailTemplateRenderer,
	NewEmailSender,

	bun.ProviderSet,
)
