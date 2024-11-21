package loadtest

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
)

var (
	_ loadtest.ClientFactory = (*MsgSendClientFactory)(nil)
	_ loadtest.Client        = (*MsgSendClientFactory)(nil)
)

type MsgSendClientFactory struct {
	*BaseInfo
	denom string
}

func NewMsgSendClientFactory(baseInfo *BaseInfo, denom string) *MsgSendClientFactory {
	baseInfo.GasLimit = 100000
	baseInfo.GasPrice = types.NewCoin(baseInfo.GetDenom(), sdkmath.NewInt(1))
	return &MsgSendClientFactory{BaseInfo: baseInfo, denom: denom}
}

func (c *MsgSendClientFactory) Name() string {
	return "msg_send"
}

func (c *MsgSendClientFactory) ValidateConfig(cfg loadtest.Config) error {
	return nil
}

func (c *MsgSendClientFactory) NewClient(cfg loadtest.Config) (loadtest.Client, error) {
	c.Memo = fmt.Sprintf("msg_send_%d", cfg.Rate)
	return c, nil
}

func (c *MsgSendClientFactory) GenerateTx() ([]byte, error) {
	account := c.Accounts.NextAccount()
	msgs := []types.Msg{&banktypes.MsgSend{
		FromAddress: account.Address,
		ToAddress:   account.Address,
		Amount:      types.NewCoins(types.NewCoin(c.denom, sdkmath.NewInt(1))),
	}}
	return c.BuildTx(account, msgs)
}
