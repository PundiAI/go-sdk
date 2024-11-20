package telemetry_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/telemetry"
)

func TestNewServer(t *testing.T) {
	t.Skip()
	config := telemetry.NewDefConfig()
	config.ServiceName = "test"
	server := telemetry.NewServer(log.NewNopLogger(), config)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	err := server.Start(ctx, group)
	require.NoError(t, err)

	emitMetrics()

	gr, err := server.Gather(telemetry.FormatPrometheus)
	require.NoError(t, err)
	require.Equal(t, gr.ContentType, string(expfmt.FmtText))

	t.Log(string(gr.Metrics))
	// todo check metrics
	require.True(t, strings.Contains(string(gr.Metrics), "test_dummy_counter 2"))

	<-ctx.Done()
	assert.NoError(t, server.Close())
	assert.NoError(t, group.Wait())
}

func emitMetrics() {
	ticker := time.NewTicker(time.Millisecond)
	timeout := time.After(2500 * time.Microsecond)

	for {
		select {
		case <-ticker.C:
			telemetry.IncrCounter(1.0, "dummy_counter")
		case <-timeout:
			return
		}
	}
}
