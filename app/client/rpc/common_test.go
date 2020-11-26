package rpc

import (
	"github.com/gonitro/nitro/app/registry"
)

var (
	// mock data
	testData = map[string][]*registry.App{
		"foo": {
			{
				Name:    "foo",
				Version: "1.0.0",
				Instances: []*registry.Instance{
					{
						Id:      "foo-1.0.0-123",
						Address: "localhost:9999",
						Metadata: map[string]string{
							"protocol": "rpc",
						},
					},
					{
						Id:      "foo-1.0.0-321",
						Address: "localhost:9999",
						Metadata: map[string]string{
							"protocol": "rpc",
						},
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
						Metadata: map[string]string{
							"protocol": "rpc",
						},
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
						Metadata: map[string]string{
							"protocol": "rpc",
						},
					},
				},
			},
		},
	}
)
