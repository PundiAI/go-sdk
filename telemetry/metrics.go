package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// Metrics supported format types.
const (
	FormatDefault    = ""
	FormatPrometheus = "prometheus"
	FormatText       = "text"
)

// GatherResponse is the response type of registered metrics
type GatherResponse struct {
	Metrics     []byte
	ContentType string
}

func (s *Server) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := strings.TrimSpace(r.FormValue("format"))

		gr, err := s.Gather(format)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(fmt.Appendf([]byte{}, "failed to gather metrics: %s", err))
			return
		}

		w.Header().Set("Content-Type", gr.ContentType)
		_, _ = w.Write(gr.Metrics)
	}
}

// Gather collects all registered metrics and returns a GatherResponse where the
// metrics are encoded depending on the type. Metrics are either encoded via
// Prometheus or JSON if in-memory.
func (s *Server) Gather(format string) (GatherResponse, error) {
	switch format {
	case FormatPrometheus:
		return s.gatherPrometheus()

	case FormatText:
		return s.gatherGeneric()

	case FormatDefault:
		return s.gatherGeneric()

	default:
		return GatherResponse{}, fmt.Errorf("unsupported metrics format: %s", format)
	}
}

func (*Server) gatherPrometheus() (GatherResponse, error) {
	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather prometheus metrics: %w", err)
	}

	buf := &bytes.Buffer{}
	defer buf.Reset()

	e := expfmt.NewEncoder(buf, expfmt.FmtText)
	for _, mf := range metricsFamilies {
		if err := e.Encode(mf); err != nil {
			return GatherResponse{}, fmt.Errorf("failed to encode prometheus metrics: %w", err)
		}
	}

	return GatherResponse{ContentType: string(expfmt.FmtText), Metrics: buf.Bytes()}, nil
}

func (s *Server) gatherGeneric() (GatherResponse, error) {
	summary, err := s.memSink.DisplayMetrics(nil, nil)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather in-memory metrics: %w", err)
	}

	content, err := json.Marshal(summary)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to encode in-memory metrics: %w", err)
	}

	return GatherResponse{ContentType: "application/json", Metrics: content}, nil
}
