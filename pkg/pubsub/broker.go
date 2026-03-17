package pubsub

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/lynx-go/lynx/contrib/pubsub"
	"github.com/lynx-go/x/log"
)

type Publisher interface {
	Publish(ctx context.Context, topicName TopicName, eventName EventName, data any, opts ...pubsub.PublishOption) error
}

// Broker Wrap Lynx Broker to use cloudevents as standard event format
type Broker struct {
	pubsub.Broker
}

func NewPubSub(binder pubsub.Binder) *Broker {
	broker := pubsub.NewBroker(pubsub.Options{}, []pubsub.Binder{binder})
	return &Broker{broker}
}

func NewPublisher(broker *Broker) Publisher {
	return broker
}

func (b *Broker) Publish(ctx context.Context, topicName TopicName, eventName EventName, data any, opts ...pubsub.PublishOption) error {
	log.InfoContext(ctx, "publishing event", "topicName", topicName, "eventName", eventName)
	o := &pubsub.PublishOptions{}
	for _, opt := range opts {
		opt(o)
	}
	msg, err := NewMessage(b.ID(), eventName.String(), data)
	if err != nil {
		return err
	}

	return b.Broker.Publish(context.WithoutCancel(ctx), topicName.String(), msg, opts...)
}

type HandlerFunc func(ctx context.Context, e *cloudevents.Event) error

func (b *Broker) Subscribe(topicName TopicName, eventName EventName, handlerName string, h HandlerFunc, opts ...pubsub.SubscribeOption) error {
	handler := func(ctx context.Context, msg *message.Message) error {
		event := cloudevents.NewEvent()
		if err := event.UnmarshalJSON(msg.Payload); err != nil {
			return err
		}
		if event.Type() == eventName.String() {
			return h(ctx, &event)
		}

		return nil
	}
	return b.Broker.Subscribe(topicName.String(), handlerName, handler, opts...)
}

type Handler interface {
	EventName() EventName
	TopicName() TopicName
	HandlerName() string
	HandlerFunc() HandlerFunc
}

type EventName string

func (e EventName) String() string {
	return string(e)
}

type TopicName string

func (t TopicName) String() string {
	return string(t)
}
