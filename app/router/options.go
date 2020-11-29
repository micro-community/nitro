package router

import (
	"context"

	"github.com/gonitro/nitro/app/registry"
	"github.com/gonitro/nitro/app/registry/memory"
	"github.com/gonitro/nitro/util/uuid"
)

// Options are router options
type Options struct {
	// Id is router id
	Id string
	// Address is router address
	Address string
	// Gateway is network gateway
	Gateway string
	// Network is network address
	Network string
	// Registry is the local registry
	Registry registry.Table
	// Context for additional options
	Context context.Context
	// Cache routes
	Cache bool
}

// Id sets Router Id
func Id(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

// Address sets router service address
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// Gateway sets network gateway
func Gateway(g string) Option {
	return func(o *Options) {
		o.Gateway = g
	}
}

// Network sets router network
func Network(n string) Option {
	return func(o *Options) {
		o.Network = n
	}
}

// Registry sets the local registry
func Registry(r registry.Table) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// Cache the routes
func Cache() Option {
	return func(o *Options) {
		o.Cache = true
	}
}

// DefaultOptions returns router default options
func DefaultOptions() Options {
	return Options{
		Id:       uuid.New().String(),
		Network:  DefaultNetwork,
		Registry: memory.NewTable(),
		Context:  context.Background(),
	}
}

type ReadOptions struct {
	App string
}

type ReadOption func(o *ReadOptions)

// ReadApp sets the service to read from the table
func ReadApp(s string) ReadOption {
	return func(o *ReadOptions) {
		o.App = s
	}
}
