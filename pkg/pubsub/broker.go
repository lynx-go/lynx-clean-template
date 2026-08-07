package pubsub

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/lynx-go/lynx"
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

func NewPubSub(broker pubsub.Broker) *Broker {
	return &Broker{broker}
}

func NewPublisher(broker *Broker) Publisher {
	return broker
}

// Publish 构造 cloudevents 信封并直接发布业务对象（cloudevents.Event），
// 由 Broker 按 topic 的 Marshaler 自动序列化。
func (b *Broker) Publish(ctx context.Context, topicName TopicName, eventName EventName, data any, opts ...pubsub.PublishOption) error {
	log.InfoContext(ctx, "publishing event", "topicName", topicName, "eventName", eventName)
	o := &pubsub.PublishOptions{}
	for _, opt := range opts {
		opt(o)
	}
	source := lynx.IDFromContext(ctx)
	if source == "" {
		source = "lynx"
	}
	event, err := NewEvent(source, eventName.String(), data)
	if err != nil {
		return err
	}

	return b.Broker.Publish(context.WithoutCancel(ctx), topicName.String(), event, opts...)
}

type HandlerFunc func(ctx context.Context, e *cloudevents.Event) error

// Subscribe 以强类型订阅注册 handler：框架经 topic 的 Marshaler 把消息
// Payload 解码为 *cloudevents.Event（TypedMessage[cloudevents.Event]）后
// 按 event type 过滤并调用 h。
func (b *Broker) Subscribe(topicName TopicName, eventName EventName, handlerName string, h HandlerFunc, opts ...pubsub.SubscribeOption) error {
	return pubsub.Subscribe[cloudevents.Event](b.Broker, context.Background(), topicName.String(), handlerName,
		func(ctx context.Context, tm *pubsub.TypedMessage[cloudevents.Event]) error {
			event := &tm.Payload
			if event.Type() == eventName.String() {
				return h(ctx, event)
			}
			return nil
		}, opts...)
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
