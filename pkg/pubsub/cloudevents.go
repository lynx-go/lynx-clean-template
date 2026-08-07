package pubsub

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// NewEvent 构造 cloudevents 信封（id/type/source/time + JSON data）。
// 信封本身作为业务对象发布，由 Broker 的 Marshaler 序列化。
func NewEvent(source, name string, data any) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType(name)
	event.SetSource(source)
	event.SetTime(time.Now())
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, err
	}
	return &event, nil
}

// MustNewEvent 构造 cloudevents 信封，构造失败时 panic。
func MustNewEvent(source, name string, data any) *cloudevents.Event {
	event, err := NewEvent(source, name, data)
	if err != nil {
		panic(err)
	}
	return event
}
