package registry

import (
	"os"
	"testing"

	"github.com/gonitro/nitro/app/registry/memory"
	"github.com/gonitro/nitro/app/router"
)

func routerTestSetup() router.Router {
	r := memory.NewTable()
	return NewRouter(router.Registry(r))
}

func TestRouterClose(t *testing.T) {
	r := routerTestSetup()

	if err := r.Close(); err != nil {
		t.Errorf("failed to stop router: %v", err)
	}
	if len(os.Getenv("IN_TRAVIS_CI")) == 0 {
		t.Logf("TestRouterStartStop STOPPED")
	}
}
