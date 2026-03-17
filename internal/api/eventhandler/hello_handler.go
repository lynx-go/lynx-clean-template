package eventhandler

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/lynx-go/lynx-clean-template/internal/domain/events"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	lxpubsub "github.com/lynx-go/lynx/contrib/pubsub"
	"github.com/lynx-go/x/encoding/json"
	"github.com/lynx-go/x/log"
)

type HelloHandler struct {
}

func (h *HelloHandler) TopicName() pubsub.TopicName {
	return "hello"
}

func (h *HelloHandler) Options() []lxpubsub.SubscribeOption {
	return []lxpubsub.SubscribeOption{
		lxpubsub.WithContinueOnError(),
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
var _ lxpubsub.HandlerOptions = new(HelloHandler)

func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
}
