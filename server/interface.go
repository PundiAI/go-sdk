package server

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	// Start service and keep the goroutine of the blocked
	Start(group *errgroup.Group, ctx context.Context) error
	// Close service and release resources(e.g. http connect)
	Close() error
}

type Config interface {
	Name() string
}
