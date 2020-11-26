package client

import (
	"context"
	"time"

	"github.com/gonitro/nitro/util/backoff"
)

//BackoffFunc for retry
type BackoffFunc func(ctx context.Context, req Request, attempts int) (time.Duration, error)

func exponentialBackoff(ctx context.Context, req Request, attempts int) (time.Duration, error) {
	return backoff.Do(attempts), nil
}
