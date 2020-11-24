package router

import (
	"errors"
	"math/rand"
)

var (
	// ErrNoneAvailable is returned by select when no routes were provided to select from
	ErrNoneAvailable = errors.New("none available")
)

// Selector selects a route from a pool
type Selector interface {
	// Select a route from the pool using the strategy
	Select([]string, ...SelectOption) (Next, error)
}

type SelectorOptions struct{}

type SelectOption func(o *SelectorOptions)

// Next returns the next node
type Next func() string

type Random struct{}

func (r *Random) Select(routes []string, opts ...SelectOption) (Next, error) {
	// we can't select from an empty pool of routes
	if len(routes) == 0 {
		return nil, ErrNoneAvailable
	}

	// return the next func
	return func() string {
		// if there is only one route provided we'll select it
		if len(routes) == 1 {
			return routes[0]
		}

		// select a random route from the slice
		return routes[rand.Intn(len(routes)-1)]
	}, nil
}

type RoundRobin struct{}

func (r *RoundRobin) Select(routes []string, opts ...SelectOption) (Next, error) {
	if len(routes) == 0 {
		return nil, ErrNoneAvailable
	}

	var i int

	return func() string {
		route := routes[i%len(routes)]
		// increment
		i++
		return route
	}, nil
}
