package rpc

import (
	"bytes"

	"github.com/asim/nitro/app/codec"
	"github.com/asim/nitro/app/network"
	"github.com/asim/nitro/util/buf"
)

type rpcRequest struct {
	service     string
	method      string
	endpoint    string
	contentType string
	socket      network.Socket
	codec       codec.Codec
	header      map[string]string
	body        []byte
	rawBody     interface{}
	stream      bool
	first       bool
}

type rpcMessage struct {
	event       string
	contentType string
	payload     interface{}
	header      map[string]string
	body        []byte
	codec       codec.NewCodec
}

func (r *rpcRequest) Codec() codec.Reader {
	return r.codec
}

func (r *rpcRequest) ContentType() string {
	return r.contentType
}

func (r *rpcRequest) Service() string {
	return r.service
}

func (r *rpcRequest) Method() string {
	return r.method
}

func (r *rpcRequest) Endpoint() string {
	return r.endpoint
}

func (r *rpcRequest) Header() map[string]string {
	return r.header
}

func (r *rpcRequest) Body() interface{} {
	return r.rawBody
}

func (r *rpcRequest) Read() ([]byte, error) {
	// got a body
	if r.first {
		b := r.body
		r.first = false
		return b, nil
	}

	var msg network.Message
	err := r.socket.Recv(&msg)
	if err != nil {
		return nil, err
	}
	r.header = msg.Header

	return msg.Body, nil
}

func (r *rpcRequest) Stream() bool {
	return r.stream
}

func (r *rpcMessage) ContentType() string {
	return r.contentType
}

func (r *rpcMessage) Event() string {
	return r.event
}

func (r *rpcMessage) Payload() interface{} {
	return r.payload
}

func (r *rpcMessage) Header() map[string]string {
	return r.header
}

func (r *rpcMessage) Body() []byte {
	return r.body
}

func (r *rpcMessage) Codec() codec.Reader {
	b := buf.New(bytes.NewBuffer(r.body))
	return r.codec(b)
}
