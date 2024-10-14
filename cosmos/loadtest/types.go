package loadtest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/functionx/fx-core/v8/client"
	"github.com/functionx/fx-core/v8/client/grpc"
	"github.com/pkg/errors"
)

type BaseInfo struct {
	Accounts       *Accounts
	ChainID        string
	GasPrice       types.Coin
	GasLimit       uint64
	Memo           string
	EnableSequence bool
}

func newBaseInfo(accounts *Accounts, chainId, denom string) *BaseInfo {
	return &BaseInfo{
		Accounts: accounts,
		ChainID:  chainId,
		GasPrice: types.NewCoin(denom, sdkmath.NewInt(0)),
		GasLimit: 100_000,
		Memo:     fmt.Sprintf("loadtest_%s", chainId),
	}
}

func (b *BaseInfo) GetDenom() string {
	return b.GasPrice.Denom
}

func (b *BaseInfo) BuildTx(account *Account, msgs []types.Msg) ([]byte, error) {
	if b.Accounts.IsFistAccount() {
		b.GasLimit--
	}
	txRaw, err := client.BuildTxRaw(
		b.ChainID, account.Sequence, account.AccountNumber, account.PrivKey,
		msgs, b.GasPrice, b.GasLimit, 0, b.Memo,
	)
	if err != nil {
		return nil, err
	}
	if b.EnableSequence {
		account.Sequence++
	}
	txRawData, err := proto.Marshal(txRaw)
	if err != nil {
		return nil, err
	}
	return txRawData, nil
}

func NewBaseInfo(genesisOrUrl string, keyDir string) (*BaseInfo, error) {
	if strings.HasSuffix(genesisOrUrl, "config/genesis.json") {
		return NewBaseInfoFromGenesis(genesisOrUrl, keyDir)
	}
	if strings.Contains(genesisOrUrl, "://") {
		grpcCli, err := grpc.DailClient(genesisOrUrl)
		if err != nil {
			return nil, err
		}
		return NewBaseInfoFromClient(grpcCli, keyDir)
	} else {
		return nil, errors.New("invalid base info")
	}
}

func NewBaseInfoFromClient(client RPCClient, keyDir string) (*BaseInfo, error) {
	accounts, err := NewAccounts(client, keyDir)
	if err != nil {
		return nil, err
	}
	chainID, err := client.GetChainId()
	if err != nil {
		return nil, err
	}
	supply, err := client.QuerySupply()
	if err != nil {
		return nil, err
	}
	denom := supply[0].Denom
	return newBaseInfo(accounts, chainID, denom), nil
}

func NewBaseInfoFromGenesis(genesisPath string, keyDir string) (*BaseInfo, error) {
	accounts, err := NewAccountFromGenesis(genesisPath, keyDir)
	if err != nil {
		return nil, err
	}
	genesisFile, err := os.ReadFile(genesisPath)
	if err != nil {
		return nil, err
	}
	var genesis struct {
		ChainId  string `json:"chain_id"`
		AppState struct {
			Mint struct {
				Params struct {
					MintDenom string `json:"mint_denom"`
				} `json:"params"`
			} `json:"mint"`
			Staking struct {
				Params struct {
					BondDenom string `json:"bond_denom"`
				} `json:"params"`
			} `json:"staking"`
		} `json:"app_state"`
	}
	if err = json.Unmarshal(genesisFile, &genesis); err != nil {
		return nil, err
	}
	denom := genesis.AppState.Staking.Params.BondDenom
	return newBaseInfo(accounts, genesis.ChainId, denom), nil
}
