package registry

import (
	"os"
	"testing"

	"github.com/gonitro/nitro/app/registry"
)

func TestRemove(t *testing.T) {
	services := []*registry.App{
		{
			Name:    "foo",
			Version: "1.0.0",
			Instances: []*registry.Instance{
				{
					Id:      "foo-123",
					Address: "localhost:9999",
				},
			},
		},
		{
			Name:    "foo",
			Version: "1.0.0",
			Instances: []*registry.Instance{
				{
					Id:      "foo-123",
					Address: "localhost:6666",
				},
			},
		},
	}

	servs := Remove([]*registry.App{services[0]}, []*registry.App{services[1]})
	if i := len(servs); i > 0 {
		t.Errorf("Expected 0 nodes, got %d: %+v", i, servs)
	}
	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("Apps %+v", servs)
	}
}

func TestRemoveInstances(t *testing.T) {
	services := []*registry.App{
		{
			Name:    "foo",
			Version: "1.0.0",
			Instances: []*registry.Instance{
				{
					Id:      "foo-123",
					Address: "localhost:9999",
				},
				{
					Id:      "foo-321",
					Address: "localhost:6666",
				},
			},
		},
		{
			Name:    "foo",
			Version: "1.0.0",
			Instances: []*registry.Instance{
				{
					Id:      "foo-123",
					Address: "localhost:6666",
				},
			},
		},
	}

	nodes := delInstances(services[0].Instances, services[1].Instances)
	if i := len(nodes); i != 1 {
		t.Errorf("Expected only 1 node, got %d: %+v", i, nodes)
	}
	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("Instances %+v", nodes)
	}
}
