// Package memory provides a memory event
package memory

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/asim/nitro/app/event"
	maddr "github.com/asim/nitro/util/addr"
	mnet "github.com/asim/nitro/util/net"
	"github.com/asim/nitro/util/uuid"
)

type memoryBroker struct {
	opts event.Options

	addr string
	sync.RWMutex
	connected   bool
	Subscribers map[string][]*memorySubscriber
}

type memorySubscriber struct {
	id      string
	event   string
	exit    chan bool
	handler event.Handler
	opts    event.SubscribeOptions
}

func (m *memoryBroker) Options() event.Options {
	return m.opts
}

func (m *memoryBroker) Address() string {
	return m.addr
}

func (m *memoryBroker) Connect() error {
	m.Lock()
	defer m.Unlock()

	if m.connected {
		return nil
	}

	// use 127.0.0.1 to avoid scan of all network interfaces
	addr, err := maddr.Extract("127.0.0.1")
	if err != nil {
		return err
	}
	i := rand.Intn(20000)
	// set addr with port
	addr = mnet.HostPort(addr, 10000+i)

	m.addr = addr
	m.connected = true

	return nil
}

func (m *memoryBroker) Disconnect() error {
	m.Lock()
	defer m.Unlock()

	if !m.connected {
		return nil
	}

	m.connected = false

	return nil
}

func (m *memoryBroker) Init(opts ...event.Option) error {
	for _, o := range opts {
		o(&m.opts)
	}
	return nil
}

func (m *memoryBroker) Publish(ev string, msg *event.Message, opts ...event.PublishOption) error {
	m.RLock()
	if !m.connected {
		m.RUnlock()
		return errors.New("not connected")
	}

	subs, ok := m.Subscribers[ev]
	m.RUnlock()
	if !ok {
		return nil
	}

	for _, sub := range subs {
		if err := sub.handler(msg); err != nil {
			if eh := sub.opts.ErrorHandler; eh != nil {
				eh(msg, err)
			}
			continue
		}
	}

	return nil
}

func (m *memoryBroker) Subscribe(ev string, handler event.Handler, opts ...event.SubscribeOption) (event.Subscriber, error) {
	m.RLock()
	if !m.connected {
		m.RUnlock()
		return nil, errors.New("not connected")
	}
	m.RUnlock()

	var options event.SubscribeOptions
	for _, o := range opts {
		o(&options)
	}

	sub := &memorySubscriber{
		exit:    make(chan bool, 1),
		id:      uuid.New().String(),
		event:   ev,
		handler: handler,
		opts:    options,
	}

	m.Lock()
	m.Subscribers[ev] = append(m.Subscribers[ev], sub)
	m.Unlock()

	go func() {
		<-sub.exit
		m.Lock()
		var newSubscribers []*memorySubscriber
		for _, sb := range m.Subscribers[ev] {
			if sb.id == sub.id {
				continue
			}
			newSubscribers = append(newSubscribers, sb)
		}
		m.Subscribers[ev] = newSubscribers
		m.Unlock()
	}()

	return sub, nil
}

func (m *memoryBroker) String() string {
	return "memory"
}

func (m *memorySubscriber) Options() event.SubscribeOptions {
	return m.opts
}

func (m *memorySubscriber) Event() string {
	return m.event
}

func (m *memorySubscriber) Unsubscribe() error {
	m.exit <- true
	return nil
}

func NewBroker(opts ...event.Option) event.Broker {
	options := event.Options{
		Context: context.Background(),
	}

	rand.Seed(time.Now().UnixNano())
	for _, o := range opts {
		o(&options)
	}

	return &memoryBroker{
		opts:        options,
		Subscribers: make(map[string][]*memorySubscriber),
	}
}
