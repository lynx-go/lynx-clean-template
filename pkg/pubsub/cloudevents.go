package pubsub

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// NewEventBytes builds a CloudEvents JSON payload for the given event name and data.
func NewEventBytes(brokerID, name string, data any) ([]byte, error) {
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType(name)
	event.SetSource(brokerID)
	event.SetTime(time.Now())
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, err
	}
	return event.MarshalJSON()
}
