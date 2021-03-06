package rpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/gonitro/nitro/app/client"
	"github.com/gonitro/nitro/app/errors"
	"github.com/gonitro/nitro/app/registry"
	"github.com/gonitro/nitro/app/registry/memory"
	"github.com/gonitro/nitro/app/router"
	regRouter "github.com/gonitro/nitro/app/router/registry"
)

func newTestRouter() router.Router {
	reg := memory.NewTable(memory.Apps(testData))
	return regRouter.NewRouter(router.Registry(reg))
}

func TestCallAddress(t *testing.T) {
	var called bool
	service := "test.service"
	endpoint := "Test.Endpoint"
	address := "10.1.10.1:8080"

	wrap := func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			called = true

			if req.App() != service {
				return fmt.Errorf("expected service: %s got %s", service, req.App())
			}

			if req.Endpoint() != endpoint {
				return fmt.Errorf("expected service: %s got %s", endpoint, req.Endpoint())
			}

			if node != address {
				return fmt.Errorf("expected address: %s got %s", address, node)
			}

			// don't do the call
			return nil
		}
	}

	r := newTestRouter()

	c := NewClient(
		client.Router(r),
		client.WrapCall(wrap),
	)

	req := c.NewRequest(service, endpoint, nil)

	// test calling remote address
	if err := c.Call(context.Background(), req, nil, client.WithAddress(address)); err != nil {
		t.Fatal("call with address error", err)
	}

	if !called {
		t.Fatal("wrapper not called")
	}

}

func TestCallRetry(t *testing.T) {
	service := "test.service"
	endpoint := "Test.Endpoint"
	address := "10.1.10.1"

	var called int

	wrap := func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			called++
			if called == 1 {
				return errors.InternalServerError("test.error", "retry request")
			}

			// don't do the call
			return nil
		}
	}

	r := newTestRouter()
	c := NewClient(
		client.Router(r),
		client.WrapCall(wrap),
	)

	req := c.NewRequest(service, endpoint, nil)

	// test calling remote address
	if err := c.Call(context.Background(), req, nil, client.WithAddress(address)); err != nil {
		t.Fatal("call with address error", err)
	}

	// num calls
	if called < c.Options().CallOptions.Retries+1 {
		t.Fatal("request not retried")
	}
}

func TestCallWrapper(t *testing.T) {
	var called bool
	id := "test.1"
	service := "test.service"
	endpoint := "Test.Endpoint"
	address := "10.1.10.1:8080"

	wrap := func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			called = true

			if req.App() != service {
				return fmt.Errorf("expected service: %s got %s", service, req.App())
			}

			if req.Endpoint() != endpoint {
				return fmt.Errorf("expected service: %s got %s", endpoint, req.Endpoint())
			}

			if node != address {
				return fmt.Errorf("expected address: %s got %s", address, node)
			}

			// don't do the call
			return nil
		}
	}

	r := newTestRouter()
	c := NewClient(
		client.Router(r),
		client.WrapCall(wrap),
	)

	r.Options().Registry.Add(&registry.App{
		Name:    service,
		Version: "latest",
		Instances: []*registry.Instance{
			{
				Id:      id,
				Address: address,
			},
		},
	})

	req := c.NewRequest(service, endpoint, nil)
	if err := c.Call(context.Background(), req, nil); err != nil {
		t.Fatal("call wrapper error", err)
	}

	if !called {
		t.Fatal("wrapper not called")
	}
}
