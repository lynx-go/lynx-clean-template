package pubsub

import (
	"context"

	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx/contrib/pubsub"
	"github.com/lynx-go/x/log"
)

type Router struct {
	handlers  []Handler
	pubSub    *Broker
	ctx       context.Context
	cancelCtx context.CancelFunc
}

func (r *Router) Name() string {
	return "pubsub-router"
}

func (r *Router) Init(app lynx.Lynx) error {
	r.ctx, r.cancelCtx = context.WithCancel(app.Context())
	return nil
}

func (r *Router) Start(ctx context.Context) error {
	if err := r.run(ctx); err != nil {
		return err
	}
	<-r.ctx.Done()
	return nil
}

func (r *Router) Stop(ctx context.Context) {
	r.cancelCtx()
}

func NewRouter(pubSub *Broker, handlers []Handler) *Router {
	return &Router{
		pubSub:   pubSub,
		handlers: handlers,
	}
}

func (r *Router) run(ctx context.Context) error {
	for i := range r.handlers {
		h := r.handlers[i]
		ctx := log.WithContext(ctx, "handler_name", h.HandlerName(), "event_name", h.EventName())
		log.InfoContext(ctx, "binding handler")

		var opts []pubsub.SubscribeOption
		if o, ok := h.(pubsub.HandlerOptions); ok {
			opts = append(opts, o.Options()...)
		}
		if err := r.pubSub.Subscribe(h.TopicName(), h.EventName(), h.HandlerName(), h.HandlerFunc(), opts...); err != nil {
			return err
		}
	}
	return nil
}

var _ lynx.Component = new(Router)
