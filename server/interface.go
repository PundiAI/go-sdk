package server

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	// Start service and keep the goroutine of the blocked
	Start(ctx context.Context, group *errgroup.Group) error
	// Close service and release resources(e.g. http connect)
	Close() error
}

type Config interface {
	Name() string
}
