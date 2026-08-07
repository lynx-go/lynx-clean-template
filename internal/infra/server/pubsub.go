package server

import (
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/api/eventhandler"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/lynx/contrib/kafka"
	lxpubsub "github.com/lynx-go/lynx/contrib/pubsub"
)

func NewPubSub(kafkaT *kafka.Transport) *pubsub.Broker {
	// memory 作为默认回退：未配置 kafka 的逻辑 topic 走进程内 transport。
	memT := lxpubsub.NewMemoryTransport()
	transports := []lxpubsub.Transport{memT}
	if kafkaT != nil {
		transports = append(transports, kafkaT)
	}
	broker := lxpubsub.NewBroker(lxpubsub.Options{
		Transports:       transports,
		DefaultTransport: memT,
		// 消息收发 debug 级日志（发布/订阅两侧独立，经应用日志级别过滤）
		LogMessage: &lxpubsub.LogMessageOptions{
			Publish:   true,
			Subscribe: true,
		},
	})
	return pubsub.NewPubSub(broker)
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

func NewKafkaTransportForServer(config *config.AppConfig) (*kafka.Transport, error) {
	return NewKafkaTransport(config, false)
}

func NewKafkaTransportForCLI(config *config.AppConfig) (*kafka.Transport, error) {
	return NewKafkaTransport(config, true)
}

// NewKafkaTransport 从 AppConfig 的 pubsub.kafka 段构建 kafka Transport。
// 未配置任何 topic 时返回 (nil, nil)，表示 Kafka 未启用。
func NewKafkaTransport(config *config.AppConfig, forCli bool) (*kafka.Transport, error) {
	topics := map[string]kafka.TopicOptions{}
	if config.Pubsub != nil && config.Pubsub.Kafka != nil {
		for name, c := range config.Pubsub.Kafka {
			to := kafka.TopicOptions{
				Brokers: c.Brokers,
				Topics:  []string{c.Topic},
			}
			if c.Consumer != nil && !forCli {
				to.Consumer = &kafka.ConsumerOptions{
					GroupID:    c.Consumer.GroupId,
					Instances:  int(c.Consumer.Instances),
					LogMessage: c.Consumer.LogMessage,
				}
			}
			if c.Producer != nil {
				var batchSize int
				var flushFrequency time.Duration
				// CLI 模式立即发送
				if forCli {
					batchSize = 1
					flushFrequency = time.Millisecond
				} else {
					batchSize = int(c.Producer.BatchSize)
					flushFrequency, _ = time.ParseDuration(c.Producer.BatchTimeout)
				}
				to.Producer = &kafka.ProducerOptions{
					Topic:          c.Topic,
					LogMessage:     c.Producer.LogMessage,
					BatchSize:      batchSize,
					FlushFrequency: flushFrequency,
				}
			}
			topics[name] = to
		}
	}
	if len(topics) == 0 {
		return nil, nil
	}
	return kafka.NewTransport(kafka.Options{Topics: topics})
}
