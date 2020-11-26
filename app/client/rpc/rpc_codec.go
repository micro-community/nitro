package rpc

import (
	"bytes"
	errs "errors"

	"github.com/gonitro/nitro/app/codec"
	raw "github.com/gonitro/nitro/app/codec/bytes"
	"github.com/gonitro/nitro/app/codec/json"
	"github.com/gonitro/nitro/app/codec/jsonrpc"
	"github.com/gonitro/nitro/app/errors"
	"github.com/gonitro/nitro/app/network"
	"github.com/gonitro/nitro/app/registry"
)

const (
	lastStreamResponseError = "EOS"
)

// serverError represents an error that has been returned from
// the remote side of the RPC connection.
type serverError string

func (e serverError) Error() string {
	return string(e)
}

// errShutdown holds the specific error for closing/closed connections
var (
	errShutdown = errs.New("connection is shut down")
)

type rpcCodec struct {
	client network.Client
	codec  codec.Codec

	req *network.Message
	buf *readWriteCloser

	// signify if its a stream
	stream string
}

type readWriteCloser struct {
	wbuf *bytes.Buffer
	rbuf *bytes.Buffer
}

var (
	DefaultContentType = "application/json"

	DefaultCodecs = map[string]codec.NewCodec{
		"application/json":         json.NewCodec,
		"application/json-rpc":     jsonrpc.NewCodec,
		"application/octet-stream": raw.NewCodec,
	}

	// TODO: remove legacy codec list
	defaultCodecs = map[string]codec.NewCodec{
		"application/json":     jsonrpc.NewCodec,
		"application/json-rpc": jsonrpc.NewCodec,
	}
)

func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
	return rwc.rbuf.Read(p)
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
	return rwc.wbuf.Write(p)
}

func (rwc *readWriteCloser) Close() error {
	rwc.rbuf.Reset()
	rwc.wbuf.Reset()
	return nil
}

func getHeaders(m *codec.Message) {
	set := func(v, hdr string) string {
		if len(v) > 0 {
			return v
		}
		return m.Header[hdr]
	}

	// check error in header
	m.Error = set(m.Error, "Error")

	// check endpoint in header
	m.Endpoint = set(m.Endpoint, "Endpoint")

	// check method in header
	m.Method = set(m.Method, "Method")

	// set the request id
	m.Id = set(m.Id, "Id")
}

func setHeaders(m *codec.Message, stream string) {
	set := func(hdr, v string) {
		if len(v) == 0 {
			return
		}
		m.Header[hdr] = v
	}

	set("Id", m.Id)
	set("Service", m.Target)
	set("Method", m.Method)
	set("Endpoint", m.Endpoint)
	set("Error", m.Error)

	if len(stream) > 0 {
		set("Stream", stream)
	}
}

// setupProtocol sets up the old protocol
func setupProtocol(msg *network.Message, node *registry.Node) codec.NewCodec {
	// get the protocol from node metadata
	if protocol := node.Metadata["protocol"]; len(protocol) > 0 {
		return nil
	}

	// processing event publishing
	if len(msg.Header["Event"]) > 0 {
		return nil
	}

	// no protocol use old codecs
	switch msg.Header["Content-Type"] {
	case "application/json":
		msg.Header["Content-Type"] = "application/json-rpc"
	}

	// now return codec
	return defaultCodecs[msg.Header["Content-Type"]]
}

func newRpcCodec(req *network.Message, client network.Client, c codec.NewCodec, stream string) codec.Codec {
	rwc := &readWriteCloser{
		wbuf: bytes.NewBuffer(nil),
		rbuf: bytes.NewBuffer(nil),
	}
	r := &rpcCodec{
		buf:    rwc,
		client: client,
		codec:  c(rwc),
		req:    req,
		stream: stream,
	}
	return r
}

func (c *rpcCodec) Write(m *codec.Message, body interface{}) error {
	c.buf.wbuf.Reset()

	// create header
	if m.Header == nil {
		m.Header = map[string]string{}
	}

	// copy original header
	for k, v := range c.req.Header {
		m.Header[k] = v
	}

	// set the mucp headers
	setHeaders(m, c.stream)

	// if body is bytes Frame don't encode
	if body != nil {
		if b, ok := body.(*raw.Frame); ok {
			// set body
			m.Body = b.Data
		} else {
			// write to codec
			if err := c.codec.Write(m, body); err != nil {
				return errors.InternalServerError("nitro.codec", err.Error())
			}
			// set body
			m.Body = c.buf.wbuf.Bytes()
		}
	}

	// create new network message
	msg := network.Message{
		Header: m.Header,
		Body:   m.Body,
	}

	// send the request
	if err := c.client.Send(&msg); err != nil {
		return errors.InternalServerError("nitro.network", err.Error())
	}

	return nil
}

func (c *rpcCodec) ReadHeader(m *codec.Message, r codec.MessageType) error {
	var tm network.Message

	// read message from network
	if err := c.client.Recv(&tm); err != nil {
		return errors.InternalServerError("nitro.network", err.Error())
	}

	c.buf.rbuf.Reset()
	c.buf.rbuf.Write(tm.Body)

	// set headers from network
	m.Header = tm.Header

	// read header
	err := c.codec.ReadHeader(m, r)

	// get headers
	getHeaders(m)

	// return header error
	if err != nil {
		return errors.InternalServerError("nitro.codec", err.Error())
	}

	return nil
}

func (c *rpcCodec) ReadBody(b interface{}) error {
	// read body
	// read raw data
	if v, ok := b.(*raw.Frame); ok {
		v.Data = c.buf.rbuf.Bytes()
		return nil
	}

	if err := c.codec.ReadBody(b); err != nil {
		return errors.InternalServerError("nitro.codec", err.Error())
	}
	return nil
}

func (c *rpcCodec) Close() error {
	c.buf.Close()
	c.codec.Close()
	if err := c.client.Close(); err != nil {
		return errors.InternalServerError("nitro.network", err.Error())
	}
	return nil
}

func (c *rpcCodec) String() string {
	return "rpc"
}
