package shared

// DomainEvent is a marker interface for all domain events.
// Events are raised by aggregate roots to signal that something
// significant happened in the domain. They are published by the
// application layer after the domain operation commits successfully.
type DomainEvent interface {
	// TopicName returns the message broker topic for this event.
	TopicName() string
	// EventName returns the specific event identifier.
	EventName() string
}

// AggregateRoot is an embeddable base type for domain aggregate roots.
// Embed it in the primary entity of each aggregate to gain domain event
// collection support.
//
//	type User struct {
//	    shared.AggregateRoot
//	    ID   idgen.ID
//	    ...
//	}
//
// The application layer drains events after a successful commit:
//
//	events := user.DomainEvents()
//	user.ClearDomainEvents()
//	for _, e := range events { publisher.Publish(ctx, e) }
type AggregateRoot struct {
	domainEvents []DomainEvent
}

// AddDomainEvent appends an event to the aggregate's pending event list.
func (ar *AggregateRoot) AddDomainEvent(e DomainEvent) {
	ar.domainEvents = append(ar.domainEvents, e)
}

// DomainEvents returns the pending domain events collected since the last clear.
func (ar *AggregateRoot) DomainEvents() []DomainEvent {
	return ar.domainEvents
}

// ClearDomainEvents removes all pending domain events.
// Call this after the application layer has published the events.
func (ar *AggregateRoot) ClearDomainEvents() {
	ar.domainEvents = nil
}
