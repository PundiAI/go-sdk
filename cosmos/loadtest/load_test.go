package loadtest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/informalsystems/tm-load-test/pkg/loadtest"
)

func Test_LoadTest(t *testing.T) {
	t.Skip("skip load test")

	chainName := "sei"
	genesisFilePath := filepath.Join(os.ExpandEnv("$HOME"), "."+chainName, "config", "genesis.json")
	keyDir := filepath.Join(os.ExpandEnv("$HOME"), chainName+"_test_accounts")
	baseInfo, err := NewBaseInfoFromGenesis(genesisFilePath, keyDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("init accounts success", baseInfo.Accounts.Len())

	msgSendClientFactory := NewMsgSendClientFactory(baseInfo, baseInfo.GetDenom())
	if err = loadtest.RegisterClientFactory(msgSendClientFactory.Name(), msgSendClientFactory); err != nil {
		t.Fatal(err)
	}

	cfg := loadtest.Config{
		ClientFactory:     msgSendClientFactory.Name(),
		Connections:       1,
		Time:              120,
		SendPeriod:        1,
		Rate:              200,
		Count:             -1,
		BroadcastTxMethod: "async",
		Endpoints: []string{
			"ws://127.0.0.1:26657/websocket",
		},
		EndpointSelectMethod: loadtest.SelectSuppliedEndpoints,
		ExpectPeers:          0,
		MaxEndpoints:         0,
		MinConnectivity:      0,
		PeerConnectTimeout:   600,
		NoTrapInterrupts:     false,
	}
	cfg.StatsOutputFile = fmt.Sprintf("out/stats_%s_%d_%d.csv", cfg.ClientFactory, cfg.Time, cfg.Rate)

	if err = loadtest.ExecuteStandalone(cfg); err != nil {
		t.Fatal(err)
	}
}
