package pubsub

import (
	"context"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx/eventbus"
	"github.com/lynx-go/x/log"
)

// Router is a Lynx service that binds handlers to the event bus on start.
type Router struct {
	broker    *Broker
	handlers  []Handler
	ctx       context.Context
	cancelCtx context.CancelFunc
}

func (r *Router) Name() string {
	return "pubsub-router"
}

func (r *Router) Init(app lynx.AppContext) error {
	r.ctx, r.cancelCtx = context.WithCancel(app.Context())
	return r.run(app.Context())
}

func (r *Router) Start(ctx context.Context) error {
	<-r.ctx.Done()
	return nil
}

func (r *Router) Stop(ctx context.Context) error {
	r.cancelCtx()
	return nil
}

func NewRouter(broker *Broker, handlers []Handler) *Router {
	return &Router{
		broker:   broker,
		handlers: handlers,
	}
}

func (r *Router) run(ctx context.Context) error {
	for i := range r.handlers {
		h := r.handlers[i]
		ctx := log.WithContext(ctx, "handler_name", h.HandlerName(), "event_name", h.EventName())
		log.InfoContext(ctx, "binding handler")

		var opts []eventbus.SubscribeOption
		if o, ok := h.(HandlerOptions); ok {
			opts = append(opts, o.Options()...)
		}
		if err := r.broker.Subscribe(h.TopicName(), h.EventName(), h.HandlerName(), h.HandlerFunc(), opts...); err != nil {
			return err
		}
	}
	return nil
}

var _ lynx.Service = new(Router)
