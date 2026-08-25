package server

import (
	"github.com/lynx-go/lynx-clean-template/internal/api/eventhandler"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/eventbus"
)

// NewPubSub builds a Broker backed by the application event bus.
func NewPubSub(bus eventbus.Bus) *pubsub.Broker {
	return pubsub.NewPubSub(bus)
}

func NewPublisher(broker *pubsub.Broker) pubsub.Publisher {
	return broker
}

func NewPubSubRouter(
	pubSub *pubsub.Broker,
	hello *eventhandler.HelloHandler,
) *pubsub.Router {
	return pubsub.NewRouter(pubSub, []pubsub.Handler{
		hello,
	})
}
