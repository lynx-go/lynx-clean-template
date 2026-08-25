package eventhandler

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/lynx-go/lynx-clean-template/internal/domain/events"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/eventbus"
	"github.com/lynx-go/x/encoding/json"
	"github.com/lynx-go/x/log"
)

type HelloHandler struct {
}

func (h *HelloHandler) TopicName() pubsub.TopicName {
	return "hello"
}

func (h *HelloHandler) Options() []eventbus.SubscribeOption {
	return []eventbus.SubscribeOption{
		eventbus.WithContinueOnError(),
	}
}

func (h *HelloHandler) EventName() pubsub.EventName {
	return "hello"
}

func (h *HelloHandler) HandlerName() string {
	return "HelloHandler"
}

func (h *HelloHandler) HandlerFunc() pubsub.HandlerFunc {
	return func(ctx context.Context, e *cloudevents.Event) error {
		data := &events.HelloEvent{}
		if err := e.DataAs(data); err != nil {
			return err
		}
		log.InfoContext(ctx, "recv hello event", "event_data", json.MustMarshalToString(data))
		return nil
	}
}

var _ pubsub.Handler = new(HelloHandler)
var _ pubsub.HandlerOptions = new(HelloHandler)

func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
}
