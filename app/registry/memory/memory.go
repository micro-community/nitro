// Package memory provides an in-memory registry
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/gonitro/nitro/app/logger"
	"github.com/gonitro/nitro/app/registry"
	"github.com/gonitro/nitro/util/uuid"
)

var (
	sendEventTime = 10 * time.Millisecond
	ttlPruneTime  = time.Second
)

type node struct {
	*registry.Instance
	TTL      time.Duration
	LastSeen time.Time
}

type record struct {
	Name      string
	Version   string
	Metadata  map[string]string
	Instances map[string]*node
	Endpoints []*registry.Endpoint
}

type Table struct {
	options registry.Options

	sync.RWMutex
	// records is a KV map with domain name as the key and a services map as the value
	records  map[string]services
	watchers map[string]*Watcher
}

// services is a KV map with service name as the key and a map of records as the value
type services map[string]map[string]*record

// NewTable returns an initialized in-memory registry
func NewTable(opts ...registry.Option) registry.Table {
	options := registry.Options{
		Context: context.Background(),
	}
	for _, o := range opts {
		o(&options)
	}

	// records can be passed for testing purposes
	records := getAppRecords(options.Context)
	if records == nil {
		records = make(services)
	}

	reg := &Table{
		options:  options,
		records:  map[string]services{registry.DefaultDomain: records},
		watchers: make(map[string]*Watcher),
	}

	go reg.ttlPrune()

	return reg
}

func (m *Table) ttlPrune() {
	prune := time.NewTicker(ttlPruneTime)
	defer prune.Stop()

	for {
		select {
		case <-prune.C:
			m.Lock()
			for domain, services := range m.records {
				for service, versions := range services {
					for version, record := range versions {
						for id, n := range record.Instances {
							if n.TTL != 0 && time.Since(n.LastSeen) > n.TTL {
								if logger.V(logger.DebugLevel, logger.DefaultLogger) {
									logger.Debugf("Table TTL expired for node %s of service %s", n.Id, service)
								}
								delete(m.records[domain][service][version].Instances, id)
							}
						}
					}
				}
			}
			m.Unlock()
		}
	}
}

func (m *Table) sendEvent(r *registry.Result) {
	m.RLock()
	watchers := make([]*Watcher, 0, len(m.watchers))
	for _, w := range m.watchers {
		watchers = append(watchers, w)
	}
	m.RUnlock()

	for _, w := range watchers {
		select {
		case <-w.exit:
			m.Lock()
			delete(m.watchers, w.id)
			m.Unlock()
		default:
			select {
			case w.res <- r:
			case <-time.After(sendEventTime):
			}
		}
	}
}

func (m *Table) Init(opts ...registry.Option) error {
	for _, o := range opts {
		o(&m.options)
	}

	// add services
	m.Lock()
	defer m.Unlock()

	// get the existing services from the records
	srvs, ok := m.records[registry.DefaultDomain]
	if !ok {
		srvs = make(services)
	}

	// loop through the services and if it doesn't yet exist, add it to the slice. This is used for
	// testing purposes.
	for name, record := range getAppRecords(m.options.Context) {
		if _, ok := srvs[name]; !ok {
			srvs[name] = record
			continue
		}

		for version, r := range record {
			if _, ok := srvs[name][version]; !ok {
				srvs[name][version] = r
				continue
			}
		}
	}

	// set the services in the registry
	m.records[registry.DefaultDomain] = srvs
	return nil
}

func (m *Table) Options() registry.Options {
	return m.options
}

func (m *Table) Add(s *registry.App, opts ...registry.AddOption) error {
	m.Lock()
	defer m.Unlock()

	// parse the options, fallback to the default domain
	var options registry.AddOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Domain) == 0 {
		options.Domain = registry.DefaultDomain
	}

	// get the services for this domain from the registry
	srvs, ok := m.records[options.Domain]
	if !ok {
		srvs = make(services)
	}

	// domain is set in metadata so it can be passed to watchers
	if s.Metadata == nil {
		s.Metadata = map[string]string{"domain": options.Domain}
	} else {
		s.Metadata["domain"] = options.Domain
	}

	// ensure the service name exists
	r := serviceToRecord(s, options.TTL)
	if _, ok := srvs[s.Name]; !ok {
		srvs[s.Name] = make(map[string]*record)
	}

	if _, ok := srvs[s.Name][s.Version]; !ok {
		srvs[s.Name][s.Version] = r
		if logger.V(logger.DebugLevel, logger.DefaultLogger) {
			logger.Debugf("Table added new service: %s, version: %s", s.Name, s.Version)
		}
		m.records[options.Domain] = srvs
		go m.sendEvent(&registry.Result{Action: "create", App: s})
	}

	var addedInstances bool

	for _, n := range s.Instances {
		// check if already exists
		if _, ok := srvs[s.Name][s.Version].Instances[n.Id]; ok {
			continue
		}

		metadata := make(map[string]string)

		// make copy of metadata
		for k, v := range n.Metadata {
			metadata[k] = v
		}

		// set the domain
		metadata["domain"] = options.Domain

		// add the node
		srvs[s.Name][s.Version].Instances[n.Id] = &node{
			Instance: &registry.Instance{
				Id:       n.Id,
				Address:  n.Address,
				Metadata: metadata,
			},
			TTL:      options.TTL,
			LastSeen: time.Now(),
		}

		addedInstances = true
	}

	if addedInstances {
		if logger.V(logger.DebugLevel, logger.DefaultLogger) {
			logger.Debugf("Table added new node to service: %s, version: %s", s.Name, s.Version)
		}
		go m.sendEvent(&registry.Result{Action: "update", App: s})
	} else {
		// refresh TTL and timestamp
		for _, n := range s.Instances {
			if logger.V(logger.DebugLevel, logger.DefaultLogger) {
				logger.Debugf("Updated registration for service: %s, version: %s", s.Name, s.Version)
			}
			srvs[s.Name][s.Version].Instances[n.Id].TTL = options.TTL
			srvs[s.Name][s.Version].Instances[n.Id].LastSeen = time.Now()
		}
	}

	m.records[options.Domain] = srvs
	return nil
}

func (m *Table) Remove(s *registry.App, opts ...registry.RemoveOption) error {
	m.Lock()
	defer m.Unlock()

	// parse the options, fallback to the default domain
	var options registry.RemoveOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Domain) == 0 {
		options.Domain = registry.DefaultDomain
	}

	// domain is set in metadata so it can be passed to watchers
	if s.Metadata == nil {
		s.Metadata = map[string]string{"domain": options.Domain}
	} else {
		s.Metadata["domain"] = options.Domain
	}

	// if the domain doesn't exist, there is nothing to deregister
	services, ok := m.records[options.Domain]
	if !ok {
		return nil
	}

	// if no services with this name and version exist, there is nothing to deregister
	versions, ok := services[s.Name]
	if !ok {
		return nil
	}

	version, ok := versions[s.Version]
	if !ok {
		return nil
	}

	// deregister all of the service nodes from this version
	for _, n := range s.Instances {
		if _, ok := version.Instances[n.Id]; ok {
			if logger.V(logger.DebugLevel, logger.DefaultLogger) {
				logger.Debugf("Table removed node from service: %s, version: %s", s.Name, s.Version)
			}
			delete(version.Instances, n.Id)
		}
	}

	// if the nodes not empty, we replace the version in the store and exist, the rest of the logic
	// is cleanup
	if len(version.Instances) > 0 {
		m.records[options.Domain][s.Name][s.Version] = version
		go m.sendEvent(&registry.Result{Action: "update", App: s})
		return nil
	}

	// if this version was the only version of the service, we can remove the whole service from the
	// registry and exit
	if len(versions) == 1 {
		delete(m.records[options.Domain], s.Name)
		go m.sendEvent(&registry.Result{Action: "delete", App: s})

		if logger.V(logger.DebugLevel, logger.DefaultLogger) {
			logger.Debugf("Table removed service: %s", s.Name)
		}
		return nil
	}

	// there are other versions of the service running, so only remove this version of it
	delete(m.records[options.Domain][s.Name], s.Version)
	go m.sendEvent(&registry.Result{Action: "delete", App: s})
	if logger.V(logger.DebugLevel, logger.DefaultLogger) {
		logger.Debugf("Table removed service: %s, version: %s", s.Name, s.Version)
	}

	return nil
}

func (m *Table) Get(name string, opts ...registry.GetOption) ([]*registry.App, error) {
	// parse the options, fallback to the default domain
	var options registry.GetOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Domain) == 0 {
		options.Domain = registry.DefaultDomain
	}

	// if it's a wildcard domain, return from all domains
	if options.Domain == registry.WildcardDomain {
		m.RLock()
		recs := m.records
		m.RUnlock()

		var services []*registry.App

		for domain := range recs {
			srvs, err := m.Get(name, append(opts, registry.GetDomain(domain))...)
			if err == registry.ErrNotFound {
				continue
			} else if err != nil {
				return nil, err
			}
			services = append(services, srvs...)
		}

		if len(services) == 0 {
			return nil, registry.ErrNotFound
		}
		return services, nil
	}

	m.RLock()
	defer m.RUnlock()

	// check the domain exists
	services, ok := m.records[options.Domain]
	if !ok {
		return nil, registry.ErrNotFound
	}

	// check the service exists
	versions, ok := services[name]
	if !ok || len(versions) == 0 {
		return nil, registry.ErrNotFound
	}

	// serialize the response
	result := make([]*registry.App, len(versions))

	var i int

	for _, r := range versions {
		result[i] = recordToApp(r, options.Domain)
		i++
	}

	return result, nil
}

func (m *Table) List(opts ...registry.ListOption) ([]*registry.App, error) {
	// parse the options, fallback to the default domain
	var options registry.ListOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.Domain) == 0 {
		options.Domain = registry.DefaultDomain
	}

	// if it's a wildcard domain, list from all domains
	if options.Domain == registry.WildcardDomain {
		m.RLock()
		recs := m.records
		m.RUnlock()

		var services []*registry.App

		for domain := range recs {
			srvs, err := m.List(append(opts, registry.ListDomain(domain))...)
			if err != nil {
				return nil, err
			}
			services = append(services, srvs...)
		}

		return services, nil
	}

	m.RLock()
	defer m.RUnlock()

	// ensure the domain exists
	services, ok := m.records[options.Domain]
	if !ok {
		return make([]*registry.App, 0), nil
	}

	// serialize the result, each version counts as an individual service
	var result []*registry.App

	for domain, service := range services {
		for _, version := range service {
			result = append(result, recordToApp(version, domain))
		}
	}

	return result, nil
}

func (m *Table) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	// parse the options, fallback to the default domain
	var wo registry.WatchOptions
	for _, o := range opts {
		o(&wo)
	}
	if len(wo.Domain) == 0 {
		wo.Domain = registry.DefaultDomain
	}

	// construct the watcher
	w := &Watcher{
		exit: make(chan bool),
		res:  make(chan *registry.Result),
		id:   uuid.New().String(),
		wo:   wo,
	}

	m.Lock()
	m.watchers[w.id] = w
	m.Unlock()

	return w, nil
}

func (m *Table) String() string {
	return "memory"
}
