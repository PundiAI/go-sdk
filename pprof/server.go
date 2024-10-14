package pprof

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/server"
)

var _ server.Server = (*Server)(nil)

type Server struct {
	logger log.Logger
	config Config

	pprof *http.Server
}

func NewServer(logger log.Logger, config Config) *Server {
	return &Server{
		logger: logger.With("server", "pprof"),
		config: config,
	}
}

func (s *Server) Start(group *errgroup.Group, ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	if err := s.config.Check(); err != nil {
		return err
	}
	s.logger.Info("init pprof server")

	s.pprof = &http.Server{
		Addr:              s.config.ListenAddr,
		ReadHeaderTimeout: s.config.ReadTimeout,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	s.logger.Info("starting pprof server", "addr", fmt.Sprintf("http://%s", s.pprof.Addr))
	s.pprof.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}
	group.Go(func() error {
		if err := s.pprof.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("pprof HTTP server listen", "error", err)
			return errors.Wrap(err, "failed to start pprof HTTP server")
		}
		return nil
	})
	return nil
}

func (s *Server) Close() error {
	if s.pprof == nil {
		return nil
	}
	s.logger.Info("closing pprof server")
	if err := s.pprof.Close(); err != nil {
		return errors.Wrap(err, "failed to close pprof HTTP server")
	}
	return nil
}
