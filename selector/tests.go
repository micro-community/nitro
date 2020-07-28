package selector

import (
	"testing"

	"github.com/micro/go-micro/v3/router"
	"github.com/stretchr/testify/assert"
)

// Tests runs all the tests against a selector to ensure the implementations are consistent
func Tests(t *testing.T, s Selector) {
	r1 := router.Route{Service: "go.micro.service.foo", Address: "127.0.0.1:8000"}
	r2 := router.Route{Service: "go.micro.service.foo", Address: "127.0.0.1:8001"}

	t.Run("Select", func(t *testing.T) {
		t.Run("NoRoutes", func(t *testing.T) {
			srv, err := s.Select([]router.Route{})
			assert.Nil(t, srv, "Route should be nil")
			assert.Equal(t, ErrNoneAvailable, err, "Expected error to be none available")
		})

		t.Run("OneRoute", func(t *testing.T) {
			srv, err := s.Select([]router.Route{r1})
			assert.Nil(t, err, "Error should be nil")
			assert.Equal(t, r1, *srv, "Expected the route to be returned")
		})

		t.Run("MultipleRoutes", func(t *testing.T) {
			srv, err := s.Select([]router.Route{r1, r2})
			assert.Nil(t, err, "Error should be nil")
			if srv.Address != r1.Address && srv.Address != r2.Address {
				t.Errorf("Expected the route to be one of the inputs")
			}
		})

		t.Run("Filters", func(t *testing.T) {
			var filterApplied bool
			filter := func(rts []router.Route) []router.Route {
				filterApplied = true
				return rts
			}

			_, err := s.Select([]router.Route{r1, r2}, WithFilter(filter))
			assert.Nil(t, err, "Error should be nil")
			assert.True(t, filterApplied, "Filters should be applied")
		})
	})

	t.Run("Record", func(t *testing.T) {
		err := s.Record(r1, nil)
		assert.Nil(t, err, "Expected the error to be nil")
	})

	t.Run("String", func(t *testing.T) {
		assert.NotEmpty(t, s.String(), "String returned a blank string")
	})
}
