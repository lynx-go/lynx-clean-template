package server

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/api/eventhandler"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/contrib/kafka"
)

func NewPubSub(binder *kafka.Binder) *pubsub.Broker {
	return pubsub.NewPubSub(binder)
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

// NewMessageLoopClient creates a MessageLoop client for publishing events
func NewMessageLoopClient(config *config.AppConfig) (messageloopgo.Client, error) {
	mlConfig := config.GetServer().GetMessageloopGrpc()
	if mlConfig == nil {
		// MessageLoop config not available - notifications won't be forwarded
		return nil, nil
	}

	addr := mlConfig.Addr
	if addr == "" {
		// MessageLoop address not configured - notifications won't be forwarded
		return nil, nil
	}

	// Connect to MessageLoop gRPC
	client, err := messageloopgo.DialGRPC(addr,
		messageloopgo.WithDialTimeout(5*time.Second),
		messageloopgo.WithToken("skyline-backend"),
	)
	if err != nil {
		// Connection failed - notifications won't be forwarded
		return nil, nil
	}

	// Start connection in background
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		// Connection failed - notifications won't be forwarded
		return nil, nil
	}

	return client, nil
}

func NewKafkaBinderForServer(config *config.AppConfig) *kafka.Binder {
	return NewKafkaBinder(config, false)
}

func NewKafkaBinderForCLI(config *config.AppConfig) *kafka.Binder {
	return NewKafkaBinder(config, true)
}

func NewKafkaBinder(config *config.AppConfig, forCli bool) *kafka.Binder {
	bindOptions := kafka.BinderOptions{
		SubscribeOptions: map[string]kafka.ConsumerOptions{},
		PublishOptions:   map[string]kafka.ProducerOptions{},
	}
	if config.Pubsub != nil && config.Pubsub.Kafka != nil {
		cfgs := config.Pubsub.Kafka
		for k, c := range cfgs {
			if c.Consumer != nil && !forCli {
				bindOptions.SubscribeOptions[k] = kafka.ConsumerOptions{
					Brokers:     c.Brokers,
					Topic:       c.Topic,
					Group:       c.Consumer.GroupId,
					Instances:   int(c.Consumer.Instances),
					LogMessage:  c.Consumer.LogMessage,
					MappedEvent: c.Consumer.MappedEvent,
				}
			}
			if c.Producer != nil {
				var batchTimeout time.Duration
				var batchSize int
				var async bool
				// CLI 模式立即发送
				if forCli {
					batchTimeout = 1 * time.Millisecond
					batchSize = 1
					async = false
				} else {
					batchSize = int(c.Producer.BatchSize)
					batchTimeout, _ = time.ParseDuration(c.Producer.BatchTimeout)
					async = c.Producer.Async
				}
				bindOptions.PublishOptions[k] = kafka.ProducerOptions{
					Brokers:      c.Brokers,
					Topic:        c.Topic,
					LogMessage:   c.Producer.LogMessage,
					MappedEvent:  c.Producer.MappedEvent,
					BatchSize:    batchSize,
					BatchTimeout: batchTimeout,
					Async:        async,
				}
			}
		}
	}

	return kafka.NewBinder(bindOptions)
}
