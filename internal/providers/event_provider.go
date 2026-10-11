package providers

import "context"

// Event is what a listener receives. Payload holds the raw JSON body
// so each listener decodes it into its own event struct.
type Event struct {
	Topic   string
	Payload []byte
}

// EventHandler processes an event. Returning an error tells the provider
// the event was not handled, so it can be retried or discarded.
type EventHandler func(ctx context.Context, event Event) error

type PublishEventInput struct {
	Topic   string
	Payload any // encoded as JSON
}

type AddListenerInput struct {
	Topic   string
	Handler EventHandler
}

type EventProvider interface {
	PublishEvent(ctx context.Context, input PublishEventInput) error
	// AddListener registers a handler for a topic. Must be called before StartListening.
	AddListener(input AddListenerInput)
	// StartListening consumes events for every registered listener and blocks
	// until ctx is cancelled (returns nil) or the connection is lost (returns the error).
	StartListening(ctx context.Context) error
}
