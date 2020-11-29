package memory

import (
	"errors"

	"github.com/gonitro/nitro/app/registry"
)

type Watcher struct {
	id   string
	wo   registry.WatchOptions
	res  chan *registry.Result
	exit chan bool
}

func (m *Watcher) Next() (*registry.Result, error) {
	for {
		select {
		case r := <-m.res:
			if r.App == nil {
				continue
			}

			if len(m.wo.App) > 0 && m.wo.App != r.App.Name {
				continue
			}

			// extract domain from service metadata
			var domain string
			if r.App.Metadata != nil && len(r.App.Metadata["domain"]) > 0 {
				domain = r.App.Metadata["domain"]
			} else {
				domain = registry.DefaultDomain
			}

			// only send the event if watching the wildcard or this specific domain
			if m.wo.Domain == registry.GlobalDomain || m.wo.Domain == domain {
				return r, nil
			}
		case <-m.exit:
			return nil, errors.New("watcher stopped")
		}
	}
}

func (m *Watcher) Stop() {
	select {
	case <-m.exit:
		return
	default:
		close(m.exit)
	}
}
