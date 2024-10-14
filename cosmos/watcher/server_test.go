package watcher

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/pundiai/go-sdk/log"
)

func TestNewServer(t *testing.T) {
	cfg := NewDefConfig()

	logger, err := log.NewLogger(log.FormatConsole, "info")
	assert.NoError(t, err)
	server := NewServer(logger, cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)

	server.client = MockRpcClient{}
	assert.NoError(t, server.Start(group, ctx))

	<-ctx.Done()
	assert.NoError(t, server.Close())
	assert.Error(t, group.Wait())
}

var _ RPCClient = (*MockRpcClient)(nil)

type MockRpcClient struct{}

func (m MockRpcClient) TxByHash(txHash string) (*types.TxResponse, error) {
	return nil, assert.AnError
}

func (m MockRpcClient) GetLatestBlock() (*cmtservice.Block, error) {
	return nil, assert.AnError
}

func (m MockRpcClient) Close() error {
	return nil
}
