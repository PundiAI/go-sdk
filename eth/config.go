package eth

import (
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	ChainId           *big.Int      `yaml:"chain_id" mapstructure:"chain_id"`
	RpcUrl            string        `yaml:"rpc_url" mapstructure:"rpc_url"`
	EnabledRequestLog bool          `yaml:"enabled_request_log" mapstructure:"enabled_request_log"`
	Timeout           time.Duration `yaml:"timeout" mapstructure:"timeout"`
}

func NewDefConfig() Config {
	return Config{
		EnabledRequestLog: false,
		Timeout:           10 * time.Second,
	}
}

func (c Config) String() string {
	return fmt.Sprintf("chainId: %s, rpcUrl: %s", c.ChainId.String(), c.RpcUrl)
}

func (c Config) Validate() error {
	if c.ChainId == nil || c.ChainId.Sign() <= 0 {
		return errors.New("chain_id is empty")
	}
	if c.RpcUrl == "" {
		return errors.New("rpc_url is empty")
	}
	if c.Timeout <= 0 || c.Timeout > 600*time.Second {
		return errors.New("timeout is invalid, should be in (0, 600)s")
	}
	return nil
}
