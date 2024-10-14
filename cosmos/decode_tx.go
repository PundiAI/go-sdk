package cosmos

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

func DecodeTx(cdc codec.BinaryCodec, txBytes []byte) (ttx *tx.Tx, err error) {
	var raw tx.TxRaw
	if err = cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body tx.TxBody
	if err = cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, err
	}

	var authInfo tx.AuthInfo
	if err = cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, err
	}

	return &tx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}
