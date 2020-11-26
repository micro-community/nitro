package memory

import (
	"fmt"
	"testing"

	"github.com/gonitro/nitro/app/event"
)

func TestMemoryBroker(t *testing.T) {
	b := NewBroker()

	if err := b.Connect(); err != nil {
		t.Fatalf("Unexpected connect error %v", err)
	}

	ev := "test"
	count := 10

	fn := func(m *event.Message) error {
		return nil
	}

	sub, err := b.Subscribe(ev, fn)
	if err != nil {
		t.Fatalf("Unexpected error subscribing %v", err)
	}

	for i := 0; i < count; i++ {
		message := &event.Message{
			Header: map[string]string{
				"foo": "bar",
				"id":  fmt.Sprintf("%d", i),
			},
			Body: []byte(`hello world`),
		}

		if err := b.Publish(ev, message); err != nil {
			t.Fatalf("Unexpected error publishing %d", i)
		}
	}

	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("Unexpected error unsubscribing from %s: %v", ev, err)
	}

	if err := b.Disconnect(); err != nil {
		t.Fatalf("Unexpected connect error %v", err)
	}
}
