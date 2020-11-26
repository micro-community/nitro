package rpc

import (
	"github.com/gonitro/nitro/app/client"
)

type message struct {
	event       string
	contentType string
	payload     interface{}
}

func newMessage(event string, payload interface{}, contentType string, opts ...client.MessageOption) client.Message {
	var options client.MessageOptions
	for _, o := range opts {
		o(&options)
	}

	if len(options.ContentType) > 0 {
		contentType = options.ContentType
	}

	return &message{
		payload:     payload,
		event:       event,
		contentType: contentType,
	}
}

func (m *message) ContentType() string {
	return m.contentType
}

func (m *message) Event() string {
	return m.event
}

func (m *message) Payload() interface{} {
	return m.payload
}
