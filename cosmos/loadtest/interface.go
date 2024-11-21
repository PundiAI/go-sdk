package loadtest

import "github.com/cosmos/cosmos-sdk/types"

type RPCClient interface {
	GetAddressPrefix() (string, error)
	QueryAccount(address string) (types.AccountI, error)
	GetChainId() (string, error)
	QuerySupply() (types.Coins, error)
}
