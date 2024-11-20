package telemetry

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/armon/go-metrics"
	metricsprom "github.com/armon/go-metrics/prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"

	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/server"
	"github.com/pundiai/go-sdk/version"
)

type metricsCollector interface {
	Enabled() bool
	Metrics()
}

var _ server.Server = (*Server)(nil)

type Server struct {
	logger log.Logger
	config Config

	prometheus *http.Server
	memSink    *metrics.InmemSink
	inMemSig   *metrics.InmemSignal

	metricsCollector []metricsCollector
}

func NewServer(logger log.Logger, config Config) *Server {
	return &Server{
		logger: logger.With("server", "telemetry"),
		config: config,
	}
}

func (s *Server) Start(ctx context.Context, group *errgroup.Group) error { //nolint:revive // cyclomatic
	if !s.config.Enabled {
		return nil
	}

	if err := s.config.Check(); err != nil {
		return err
	}
	s.logger.Info("init telemetry server")

	if numGlobalLables := len(s.config.GlobalLabels); numGlobalLables > 0 {
		parsedGlobalLabels := make([]metrics.Label, numGlobalLables)
		for i, gl := range s.config.GlobalLabels {
			parsedGlobalLabels[i] = NewLabel(gl[0], gl[1])
		}
		globalLabels = parsedGlobalLabels
	}

	s.memSink = metrics.NewInmemSink(10*time.Second, time.Minute)

	promSink, err := metricsprom.NewPrometheusSinkFrom(metricsprom.PrometheusOpts{
		Registerer: prometheus.DefaultRegisterer,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create Prometheus sink")
	}

	metricsConf := metrics.DefaultConfig(version.Name)
	metricsConf.ServiceName = s.config.ServiceName
	metricsConf.EnableHostname = false
	if _, err = metrics.NewGlobal(metricsConf, metrics.FanoutSink{s.memSink, promSink}); err != nil {
		return errors.Wrap(err, "failed to create global metrics")
	}

	s.prometheus = &http.Server{
		Addr:              s.config.ListenAddr,
		ReadHeaderTimeout: s.config.ReadTimeout,
		Handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{MaxRequestsInFlight: s.config.MaxOpenConnections},
			),
		),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	s.logger.Info("starting telemetry server", "addr", fmt.Sprintf("http://%s", s.config.ListenAddr))
	s.inMemSig = metrics.DefaultInmemSignal(s.memSink)

	emitServerInfoMetrics()

	group.Go(func() error {
		if err = s.prometheus.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// Error starting or closing listener:
			s.logger.Error("prometheus HTTP server listen", "error", err)
			return errors.Wrap(err, "failed to start Prometheus HTTP server")
		}
		return nil
	})

	// Start metrics collection
	if len(s.metricsCollector) != 0 {
		group.Go(func() error {
			// NOTE: metricsCollector panic will be recovered here and logged as error,
			// it should not affect the normal operation of the entire service.
			// the reason for the error may be target service is not enabled.
			defer func() {
				if err := recover(); err != nil {
					s.logger.Error("metric collection failed to enable", "error", err)
				}
			}()
			for range time.NewTicker(s.config.Interval).C {
				for _, metric := range s.metricsCollector {
					if metric.Enabled() {
						metric.Metrics()
					}
				}
			}
			return nil
		})
	}
	return nil
}

func (s *Server) Close() error {
	if s.prometheus == nil {
		return nil
	}
	s.logger.Info("closing telemetry server")
	if s.inMemSig != nil {
		s.inMemSig.Stop()
	}
	if err := s.prometheus.Close(); err != nil {
		return errors.Wrap(err, "failed to close telemetry server")
	}
	return nil
}

func (s *Server) RegisterMetricsCollector(collector ...metricsCollector) *Server {
	s.metricsCollector = append(s.metricsCollector, collector...)
	return s
}

func emitServerInfoMetrics() {
	var ls []metrics.Label
	if len(version.Version) > 0 {
		ls = append(ls, NewLabel("version", version.Version))
	}
	if len(version.Name) > 0 {
		ls = append(ls, NewLabel("name", version.Name))
	}
	if len(ls) == 0 {
		return
	}
	SetGaugeWithLabels([]string{"server", "info"}, 1, ls)
}
