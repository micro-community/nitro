package memory

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gonitro/nitro/app/registry"
)

var (
	testData = map[string][]*registry.App{
		"foo": {
			{
				Name:    "foo",
				Version: "1.0.0",
				Instances: []*registry.Instance{
					{
						Id:      "foo-1.0.0-123",
						Address: "localhost:9999",
					},
					{
						Id:      "foo-1.0.0-321",
						Address: "localhost:9999",
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.1",
				Instances: []*registry.Instance{
					{
						Id:      "foo-1.0.1-321",
						Address: "localhost:6666",
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.3",
				Instances: []*registry.Instance{
					{
						Id:      "foo-1.0.3-345",
						Address: "localhost:8888",
					},
				},
			},
		},
		"bar": {
			{
				Name:    "bar",
				Version: "default",
				Instances: []*registry.Instance{
					{
						Id:      "bar-1.0.0-123",
						Address: "localhost:9999",
					},
					{
						Id:      "bar-1.0.0-321",
						Address: "localhost:9999",
					},
				},
			},
			{
				Name:    "bar",
				Version: "latest",
				Instances: []*registry.Instance{
					{
						Id:      "bar-1.0.1-321",
						Address: "localhost:6666",
					},
				},
			},
		},
	}
)

func TestMemoryTable(t *testing.T) {
	m := NewTable()

	fn := func(k string, v []*registry.App) {
		services, err := m.Get(k)
		if err != nil {
			t.Errorf("Unexpected error getting service %s: %v", k, err)
		}

		if len(services) != len(v) {
			t.Errorf("Expected %d services for %s, got %d", len(v), k, len(services))
		}

		for _, service := range v {
			var seen bool
			for _, s := range services {
				if s.Version == service.Version {
					seen = true
					break
				}
			}
			if !seen {
				t.Errorf("expected to find version %s", service.Version)
			}
		}
	}

	// register data
	for _, v := range testData {
		serviceCount := 0
		for _, service := range v {
			if err := m.Add(service); err != nil {
				t.Errorf("Unexpected register error: %v", err)
			}
			serviceCount++
			// after the service has been registered we should be able to query it
			services, err := m.Get(service.Name)
			if err != nil {
				t.Errorf("Unexpected error getting service %s: %v", service.Name, err)
			}
			if len(services) != serviceCount {
				t.Errorf("Expected %d services for %s, got %d", serviceCount, service.Name, len(services))
			}
		}
	}

	// using test data
	for k, v := range testData {
		fn(k, v)
	}

	services, err := m.List()
	if err != nil {
		t.Errorf("Unexpected error when listing services: %v", err)
	}

	totalAppCount := 0
	for _, testSvc := range testData {
		for range testSvc {
			totalAppCount++
		}
	}

	if len(services) != totalAppCount {
		t.Errorf("Expected total service count: %d, got: %d", totalAppCount, len(services))
	}

	// deregister
	for _, v := range testData {
		for _, service := range v {
			if err := m.Remove(service); err != nil {
				t.Errorf("Unexpected deregister error: %v", err)
			}
		}
	}

	// after all the service nodes have been deregistered we should not get any results
	for _, v := range testData {
		for _, service := range v {
			services, err := m.Get(service.Name)
			if err != registry.ErrNotFound {
				t.Errorf("Expected error: %v, got: %v", registry.ErrNotFound, err)
			}
			if len(services) != 0 {
				t.Errorf("Expected %d services for %s, got %d", 0, service.Name, len(services))
			}
		}
	}
}

func TestMemoryTableTTL(t *testing.T) {
	m := NewTable()

	for _, v := range testData {
		for _, service := range v {
			if err := m.Add(service, registry.AddTTL(time.Millisecond)); err != nil {
				t.Fatal(err)
			}
		}
	}

	time.Sleep(ttlPruneTime * 2)

	for name := range testData {
		svcs, err := m.Get(name)
		if err != nil {
			t.Fatal(err)
		}

		for _, svc := range svcs {
			if len(svc.Instances) > 0 {
				t.Fatalf("App %q still has nodes registered", name)
			}
		}
	}
}

func TestMemoryTableTTLConcurrent(t *testing.T) {
	concurrency := 1000
	waitTime := ttlPruneTime * 2
	m := NewTable()

	for _, v := range testData {
		for _, service := range v {
			if err := m.Add(service, registry.AddTTL(waitTime/2)); err != nil {
				t.Fatal(err)
			}
		}
	}

	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("test will wait %v, then check TTL timeouts", waitTime)
	}

	errChan := make(chan error, concurrency)
	syncChan := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		go func() {
			<-syncChan
			for name := range testData {
				svcs, err := m.Get(name)
				if err != nil {
					errChan <- err
					return
				}

				for _, svc := range svcs {
					if len(svc.Instances) > 0 {
						errChan <- fmt.Errorf("App %q still has nodes registered", name)
						return
					}
				}
			}

			errChan <- nil
		}()
	}

	time.Sleep(waitTime)
	close(syncChan)

	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			t.Fatal(err)
		}
	}
}

func TestMemoryWildcard(t *testing.T) {
	m := NewTable()
	testSrv := &registry.App{Name: "foo", Version: "1.0.0"}

	if err := m.Add(testSrv, registry.AddDomain("one")); err != nil {
		t.Fatalf("Add err: %v", err)
	}
	if err := m.Add(testSrv, registry.AddDomain("two")); err != nil {
		t.Fatalf("Add err: %v", err)
	}

	if recs, err := m.List(registry.ListDomain("one")); err != nil {
		t.Errorf("List err: %v", err)
	} else if len(recs) != 1 {
		t.Errorf("Expected 1 record, got %v", len(recs))
	}

	if recs, err := m.List(registry.ListDomain("*")); err != nil {
		t.Errorf("List err: %v", err)
	} else if len(recs) != 2 {
		t.Errorf("Expected 2 records, got %v", len(recs))
	}

	if recs, err := m.Get(testSrv.Name, registry.GetDomain("one")); err != nil {
		t.Errorf("Get err: %v", err)
	} else if len(recs) != 1 {
		t.Errorf("Expected 1 record, got %v", len(recs))
	}

	if recs, err := m.Get(testSrv.Name, registry.GetDomain("*")); err != nil {
		t.Errorf("Get err: %v", err)
	} else if len(recs) != 2 {
		t.Errorf("Expected 2 records, got %v", len(recs))
	}
}
