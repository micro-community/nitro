// Package registry is an interface for service discovery
package registry

import (
	"errors"
)

const (
	// WildcardDomain indicates any domain
	WildcardDomain = "*"
	// DefaultDomain to use if none was provided in options
	DefaultDomain = "nitro"
)

var (
	// Not found error when GetApp is called
	ErrNotFound = errors.New("service not found")
	// Watcher stopped error when watcher is stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Init(...Option) error
	Options() Options
	Register(*App, ...RegisterOption) error
	Deregister(*App, ...DeregisterOption) error
	Get(string, ...GetOption) ([]*App, error)
	List(...ListOption) ([]*App, error)
	Watch(...WatchOption) (Watcher, error)
	String() string
}

type App struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []*Endpoint       `json:"endpoints"`
	Instances []*Instance       `json:"nodes"`
}

type Instance struct {
	Id       string            `json:"id"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
}

type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
}

type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}

type Option func(*Options)

type RegisterOption func(*RegisterOptions)

type WatchOption func(*WatchOptions)

type DeregisterOption func(*DeregisterOptions)

type GetOption func(*GetOptions)

type ListOption func(*ListOptions)
