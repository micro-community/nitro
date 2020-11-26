package rpc

import (
	"testing"

	"github.com/gonitro/nitro/app/client"
)

func TestRequestOptions(t *testing.T) {
	r := newRequest("service", "endpoint", nil, "application/json")
	if r.App() != "service" {
		t.Fatalf("expected 'service' got %s", r.App())
	}
	if r.Endpoint() != "endpoint" {
		t.Fatalf("expected 'endpoint' got %s", r.Endpoint())
	}
	if r.ContentType() != "application/json" {
		t.Fatalf("expected 'endpoint' got %s", r.ContentType())
	}

	r2 := newRequest("service", "endpoint", nil, "application/octet", client.WithContentType("application/octet"))
	if r2.ContentType() != "application/octet" {
		t.Fatalf("expected 'endpoint' got %s", r2.ContentType())
	}
}
