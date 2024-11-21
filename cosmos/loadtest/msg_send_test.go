package loadtest

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/functionx/fx-core/v8/client/grpc"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MsgSend(t *testing.T) {
	t.Skip("skip load test")

	grpcClient, err := grpc.DailClient("http://127.0.0.1:26657")
	require.NoError(t, err)

	chainName := "sei"
	genesisFilePath := filepath.Join(os.ExpandEnv("$HOME"), "."+chainName, "config", "genesis.json")
	keyDir := filepath.Join(os.ExpandEnv("$HOME"), chainName+"_test_accounts")

	baseInfo, err := NewBaseInfoFromGenesis(genesisFilePath, keyDir)
	if err != nil {
		t.Fatal(err)
	}
	msgSendClientFactory := NewMsgSendClientFactory(baseInfo, baseInfo.GetDenom())
	t.Logf(msgSendClientFactory.GasPrice.String())

	client, err := msgSendClientFactory.NewClient(loadtest.Config{})
	if err != nil {
		t.Fatal(err)
	}
	rawTx, err := client.GenerateTx()
	if err != nil {
		t.Fatal(err)
	}
	txResp, err := grpcClient.BroadcastTxBytes(rawTx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(txResp.Code, txResp.String())
}

// go test -bench ^Benchmark_NewMsgSendTxsAndMarshal$ -benchtime 10s -count 1 -cpu 1 -run=^$ -benchmem
func Benchmark_NewMsgSendTxsAndMarshal(b *testing.B) {
	b.Skip("skip load test")

	homeDir := os.ExpandEnv("$HOME")
	keyDir := filepath.Join(homeDir, "test_accounts")

	genesisFilePath := filepath.Join(homeDir, ".simapp", "config", "genesis.json")
	accounts, err := NewAccountFromGenesis(genesisFilePath, keyDir)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = createTestTx(accounts); err != nil {
			b.Fatal(err)
		}
	}
}

func Test_MsgSend2(t *testing.T) {
	t.Skip("skip load test")

	homeDir := os.ExpandEnv("$HOME")
	keyDir := filepath.Join(homeDir, "test_accounts")

	err := CreateGenesisAccounts("sei", 10, keyDir)
	assert.NoError(t, err)

	genesisFilePath := filepath.Join(homeDir, ".simapp", "config", "genesis.json")
	accounts, err := NewAccountFromGenesis(genesisFilePath, keyDir)
	if err != nil {
		t.Fatal(err)
	}
	txs, err := createTestTx(accounts)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeTxsToFile(txs); err != nil {
		t.Fatal(err)
	}
}

func createTestTx(accounts *Accounts) ([]string, error) {
	factory := NewMsgSendClientFactory(&BaseInfo{
		Accounts: accounts,
		ChainID:  "cosmos",
		GasPrice: types.NewCoin("stake", sdkmath.NewInt(0)),
		GasLimit: 120000,
	}, "stake")

	txCount := accounts.Len()
	txs := make([]string, txCount)
	for i := 0; i < txCount; i++ {
		txRaw, err := factory.GenerateTx()
		if err != nil {
			return nil, err
		}
		txs[i] = base64.StdEncoding.EncodeToString(txRaw)
	}

	return txs, nil
}

func writeTxsToFile(txs []string) error {
	txData, err := json.Marshal(txs)
	if err != nil {
		return err
	}
	if err = os.RemoveAll(filepath.Join(os.ExpandEnv("$HOME"), "txs.json")); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(os.ExpandEnv("$HOME"), "txs.json"), txData, 0o600); err != nil {
		return err
	}
	return nil
}
