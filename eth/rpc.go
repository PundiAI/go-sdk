package eth

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httputil"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"

	"github.com/pundiai/go-sdk/log"
)

type RPCClient interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)

	bind.ContractCaller
	bind.PendingContractCaller
	bind.ContractTransactor
	bind.ContractFilterer
	bind.DeployBackend
	bind.ContractBackend

	ethereum.ChainStateReader

	Close()
}

func NewRPCClient(ctx context.Context, logger log.Logger, config Config) (RPCClient, error) {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   config.Timeout,
	}

	if config.EnabledRequestLog {
		httpClient.Transport = newLoggerTransport(logger, http.DefaultTransport)
	}

	c, err := rpc.DialOptions(ctx, config.RpcUrl, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, errors.Wrapf(err, "dial rpc url %s", config.RpcUrl)
	}

	rpcClient := &client{
		logger: logger.With("module", "eth-rpc"),
		config: config,
		Client: ethclient.NewClient(c),
	}
	return rpcClient, rpcClient.check(ctx)
}

type client struct {
	*ethclient.Client
	logger log.Logger
	config Config
}

func (c *client) check(ctx context.Context) error {
	id, err := c.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "get chain id")
	}
	if id.Cmp(c.config.ChainId) != 0 {
		return errors.Errorf("chain id mismatch, expect %d, got %d", c.config.ChainId, id)
	}
	return nil
}

var _ http.RoundTripper = (*LoggerTransport)(nil)

type LoggerTransport struct {
	http.RoundTripper
	logger log.Logger
}

func newLoggerTransport(logger log.Logger, tripper http.RoundTripper) *LoggerTransport {
	return &LoggerTransport{
		logger:       logger.With("module", "http"),
		RoundTripper: tripper,
	}
}

func (l *LoggerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	request, err := httputil.DumpRequest(req, true)
	if err == nil {
		l.logger.Debugf("request: %s", string(request))
	}
	return l.RoundTripper.RoundTrip(req)
}
