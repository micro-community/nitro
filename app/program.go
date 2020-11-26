// Package app encapsulates the client, server and other interfaces to provide a complete dapp
package app

import (
	"context"

	"github.com/gonitro/nitro/app/client"
	rpcClient "github.com/gonitro/nitro/app/client/rpc"
	mevent "github.com/gonitro/nitro/app/event/memory"
	sock "github.com/gonitro/nitro/app/network/socket"
	"github.com/gonitro/nitro/app/registry/memory"
	"github.com/gonitro/nitro/app/router/static"
	"github.com/gonitro/nitro/app/server"
	rpcServer "github.com/gonitro/nitro/app/server/rpc"
)

// Program is an interface for distributed application programming
type Program interface {
	// Set the current application name
	Name(string)
	// Execute a function in a remote program
	Execute(prog, fn string, req, rsp interface{}) error
	// Broadcast an event to subscribers
	Broadcast(event string, msg interface{}) error
	// Register a function e.g a public Go struct/method with signature func(context.Context, *Request, *Response) error
	Register(fn interface{}) error
	// Subscribe to broadcast events. Signature is public Go func or struct with signature func(context.Context, *Message) error
	Subscribe(event string, fn interface{}) error
	// Run the application program
	Run() error
}

type nitroProgram struct {
	opts Options
}

func (s *nitroProgram) Name(name string) {
	s.opts.Server.Init(
		server.Name(name),
	)
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *nitroProgram) Init(opts ...Option) {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}
}

func (s *nitroProgram) Options() Options {
	return s.opts
}

func (s *nitroProgram) Execute(name, ep string, req, rsp interface{}) error {
	r := s.Client().NewRequest(name, ep, req)
	return s.Client().Call(context.Background(), r, rsp)
}

func (s *nitroProgram) Broadcast(event string, msg interface{}) error {
	m := s.Client().NewMessage(event, msg)
	return s.Client().Publish(context.Background(), m)
}

func (s *nitroProgram) Register(v interface{}) error {
	h := s.Server().NewHandler(v)
	return s.Server().Handle(h)
}

func (s *nitroProgram) Subscribe(event string, v interface{}) error {
	sub := s.Server().NewSubscriber(event, v)
	return s.Server().Subscribe(sub)
}

func (s *nitroProgram) Client() client.Client {
	return s.opts.Client
}

func (s *nitroProgram) Server() server.Server {
	return s.opts.Server
}

func (s *nitroProgram) String() string {
	return "rpc"
}

func (s *nitroProgram) Start() error {
	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.opts.Server.Start(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *nitroProgram) Stop() error {
	var gerr error

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	if err := s.opts.Server.Stop(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (s *nitroProgram) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// wait on context cancel
	<-s.opts.Context.Done()

	return s.Stop()
}

// New returns a new application program
func New(opts ...Option) *nitroProgram {
	b := mevent.NewBroker()
	c := rpcClient.NewClient()
	s := rpcServer.NewServer()
	r := memory.NewRegistry()
	t := sock.NewTransport()
	st := static.NewRouter()

	// set client options
	c.Init(
		client.Router(st),
		client.Broker(b),
		client.Registry(r),
		client.Transport(t),
	)

	// set server options
	s.Init(
		server.Broker(b),
		server.Registry(r),
		server.Transport(t),
	)

	// define local opts
	options := Options{
		Broker:   b,
		Client:   c,
		Server:   s,
		Registry: r,
		Context:  context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &nitroProgram{
		opts: options,
	}
}
