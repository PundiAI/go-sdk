package pprof_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/pprof"
)

func TestNewServer(t *testing.T) {
	config := pprof.NewDefConfig()
	config.ListenAddr = "localhost:6061"
	server := pprof.NewServer(log.NewNopLogger(), config)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	err := server.Start(ctx, group)
	require.NoError(t, err)

	resp, err := http.Get("http://" + config.ListenAddr + "/debug/pprof/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	<-ctx.Done()
	require.NoError(t, server.Close())
	require.NoError(t, group.Wait())
}
