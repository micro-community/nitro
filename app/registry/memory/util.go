package memory

import (
	"time"

	"github.com/gonitro/nitro/app/registry"
)

func serviceToRecord(s *registry.App, ttl time.Duration) *record {
	metadata := make(map[string]string, len(s.Metadata))
	for k, v := range s.Metadata {
		metadata[k] = v
	}

	nodes := make(map[string]*node, len(s.Instances))
	for _, n := range s.Instances {
		nodes[n.Id] = &node{
			Instance:     n,
			TTL:      ttl,
			LastSeen: time.Now(),
		}
	}

	endpoints := make([]*registry.Endpoint, len(s.Endpoints))
	for i, e := range s.Endpoints {
		endpoints[i] = e
	}

	return &record{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  metadata,
		Instances:     nodes,
		Endpoints: endpoints,
	}
}

func recordToApp(r *record, domain string) *registry.App {
	metadata := make(map[string]string, len(r.Metadata))
	for k, v := range r.Metadata {
		metadata[k] = v
	}

	// set the domain in metadata so it can be determined when a wildcard query is performed
	metadata["domain"] = domain

	endpoints := make([]*registry.Endpoint, len(r.Endpoints))
	for i, e := range r.Endpoints {
		request := new(registry.Value)
		if e.Request != nil {
			*request = *e.Request
		}
		response := new(registry.Value)
		if e.Response != nil {
			*response = *e.Response
		}

		metadata := make(map[string]string, len(e.Metadata))
		for k, v := range e.Metadata {
			metadata[k] = v
		}

		endpoints[i] = &registry.Endpoint{
			Name:     e.Name,
			Request:  request,
			Response: response,
			Metadata: metadata,
		}
	}

	nodes := make([]*registry.Instance, len(r.Instances))
	i := 0
	for _, n := range r.Instances {
		metadata := make(map[string]string, len(n.Metadata))
		for k, v := range n.Metadata {
			metadata[k] = v
		}

		nodes[i] = &registry.Instance{
			Id:       n.Id,
			Address:  n.Address,
			Metadata: metadata,
		}
		i++
	}

	return &registry.App{
		Name:      r.Name,
		Version:   r.Version,
		Metadata:  metadata,
		Endpoints: endpoints,
		Instances:     nodes,
	}
}
