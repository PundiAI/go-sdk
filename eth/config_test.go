package eth_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/pundiai/go-sdk/eth"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestNewDefConfig() {
	config := eth.NewDefConfig()
	config.ChainId = big.NewInt(0)
	suite.Equal("chainId: 0, rpcUrl: ", config.String())
}

func (suite *ConfigTestSuite) TestCheck() {
	config := eth.NewDefConfig()
	suite.Require().EqualError(config.Check(), "check: chain_id is empty")

	config.ChainId = big.NewInt(1)
	config.RpcUrl = ""
	suite.Require().EqualError(config.Check(), "check: rpc_url is empty")
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
