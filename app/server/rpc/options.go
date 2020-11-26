package rpc

import (
	"github.com/gonitro/nitro/app/codec"
	mevent "github.com/gonitro/nitro/app/event/memory"
	tmem "github.com/gonitro/nitro/app/network/memory"
	"github.com/gonitro/nitro/app/registry/memory"
	"github.com/gonitro/nitro/app/server"
)

func newOptions(opt ...server.Option) server.Options {
	opts := server.Options{
		Codecs:      make(map[string]codec.NewCodec),
		Metadata:    map[string]string{},
		AddInterval: server.DefaultAddInterval,
		AddTTL:      server.DefaultAddTTL,
	}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Broker == nil {
		opts.Broker = mevent.NewBroker()
	}

	if opts.Registry == nil {
		opts.Registry = memory.NewRegistry()
	}

	if opts.Transport == nil {
		opts.Transport = tmem.NewTransport()
	}

	if opts.AddCheck == nil {
		opts.AddCheck = server.DefaultAddCheck
	}

	if len(opts.Address) == 0 {
		opts.Address = server.DefaultAddress
	}

	if len(opts.Name) == 0 {
		opts.Name = server.DefaultName
	}

	if len(opts.Id) == 0 {
		opts.Id = server.DefaultId
	}

	if len(opts.Version) == 0 {
		opts.Version = server.DefaultVersion
	}

	return opts
}
