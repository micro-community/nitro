// Package cache is a caching interface
package cache

// Cache is an interface for caching
type Cache interface {
	// Get a value
	Get(key string) (interface{}, error)
	// Set a value
	Set(key string, val interface{}) error
	// Delete a value
	Delete(key string) error
	// Name of the implementation
	String() string
}

//Options for cache
type Options struct {
	Nodes []string
}

//Option to set Options
type Option func(o *Options)

// Nodes sets the nodes for the cache
func Nodes(v ...string) Option {
	return func(o *Options) {
		o.Nodes = v
	}
}
