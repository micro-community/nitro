package memory

import (
	"context"

	"github.com/gonitro/nitro/app/registry"
)

type servicesKey struct{}

func getAppRecords(ctx context.Context) map[string]map[string]*record {
	memApps, ok := ctx.Value(servicesKey{}).(map[string][]*registry.App)
	if !ok {
		return nil
	}

	services := make(map[string]map[string]*record)

	for name, svc := range memApps {
		if _, ok := services[name]; !ok {
			services[name] = make(map[string]*record)
		}
		// go through every version of the service
		for _, s := range svc {
			services[s.Name][s.Version] = serviceToRecord(s, 0)
		}
	}

	return services
}

// Apps is an option that preloads service data
func Apps(s map[string][]*registry.App) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, servicesKey{}, s)
	}
}
