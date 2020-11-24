// Package event is an interface used for asynchronous messaging
package event

// Broker is an interface used for asynchronous messaging.
type Broker interface {
	Init(...Option) error
	Options() Options
	Address() string
	Connect() error
	Disconnect() error
	Publish(event string, m *Message, opts ...PublishOption) error
	Subscribe(event string, h Handler, opts ...SubscribeOption) (Subscriber, error)
	String() string
}

// Handler is used to process messages via a subscription of a event.
type Handler func(*Message) error

type ErrorHandler func(*Message, error)

type Message struct {
	Header map[string]string
	Body   []byte
}

// Subscriber is a convenience return type for the Subscribe method
type Subscriber interface {
	Event() string
	Options() SubscribeOptions
	Unsubscribe() error
}
