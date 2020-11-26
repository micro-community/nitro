package registry

import (
	"github.com/gonitro/nitro/app/registry"
)

func addInstances(old, neu []*registry.Instance) []*registry.Instance {
	nodes := make([]*registry.Instance, len(neu))
	// add all new nodes
	for i, n := range neu {
		node := *n
		nodes[i] = &node
	}

	// look at old nodes
	for _, o := range old {
		var exists bool

		// check against new nodes
		for _, n := range nodes {
			// ids match then skip
			if o.Id == n.Id {
				exists = true
				break
			}
		}

		// keep old node
		if !exists {
			node := *o
			nodes = append(nodes, &node)
		}
	}

	return nodes
}

func delInstances(old, del []*registry.Instance) []*registry.Instance {
	var nodes []*registry.Instance
	for _, o := range old {
		var rem bool
		for _, n := range del {
			if o.Id == n.Id {
				rem = true
				break
			}
		}
		if !rem {
			nodes = append(nodes, o)
		}
	}
	return nodes
}

// CopyApp make a copy of service
func CopyApp(service *registry.App) *registry.App {
	// copy service
	s := new(registry.App)
	*s = *service

	// copy nodes
	nodes := make([]*registry.Instance, len(service.Instances))
	for j, node := range service.Instances {
		n := new(registry.Instance)
		*n = *node
		nodes[j] = n
	}
	s.Instances = nodes

	// copy endpoints
	eps := make([]*registry.Endpoint, len(service.Endpoints))
	for j, ep := range service.Endpoints {
		e := new(registry.Endpoint)
		*e = *ep
		eps[j] = e
	}
	s.Endpoints = eps
	return s
}

// Copy makes a copy of services
func Copy(current []*registry.App) []*registry.App {
	services := make([]*registry.App, len(current))
	for i, service := range current {
		services[i] = CopyApp(service)
	}
	return services
}

// Merge merges two lists of services and returns a new copy
func Merge(olist []*registry.App, nlist []*registry.App) []*registry.App {
	var srv []*registry.App

	for _, n := range nlist {
		var seen bool
		for _, o := range olist {
			if o.Version == n.Version {
				sp := new(registry.App)
				// make copy
				*sp = *o
				// set nodes
				sp.Instances = addInstances(o.Instances, n.Instances)

				// mark as seen
				seen = true
				srv = append(srv, sp)
				break
			} else {
				sp := new(registry.App)
				// make copy
				*sp = *o
				srv = append(srv, sp)
			}
		}
		if !seen {
			srv = append(srv, Copy([]*registry.App{n})...)
		}
	}
	return srv
}

// Remove removes services and returns a new copy
func Remove(old, del []*registry.App) []*registry.App {
	var services []*registry.App

	for _, o := range old {
		srv := new(registry.App)
		*srv = *o

		var rem bool

		for _, s := range del {
			if srv.Version == s.Version {
				srv.Instances = delInstances(srv.Instances, s.Instances)

				if len(srv.Instances) == 0 {
					rem = true
				}
			}
		}

		if !rem {
			services = append(services, srv)
		}
	}

	return services
}
