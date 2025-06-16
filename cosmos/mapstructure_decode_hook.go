package cosmos

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/mitchellh/mapstructure"
)

// CustomDecodeHook is a custom decode hook for mapstructure
func CustomDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		func(f reflect.Type, t reflect.Type, data any) (any, error) {
			if f.Kind() != reflect.String {
				return data, nil
			}
			if t != reflect.TypeOf(types.Coin{}) {
				return data, nil
			}
			return types.ParseCoinNormalized(data.(string))
		},
		func(f reflect.Type, t reflect.Type, data any) (any, error) {
			if f.Kind() != reflect.String {
				return data, nil
			}
			if t != reflect.TypeOf(tx.BroadcastMode(0)) {
				return data, nil
			}
			return tx.BroadcastMode_value[data.(string)], nil
		},
		func(f reflect.Type, t reflect.Type, data any) (any, error) {
			if f.Kind() != reflect.String {
				return data, nil
			}
			if t != reflect.TypeOf(types.DecCoin{}) {
				return data, nil
			}
			return types.ParseDecCoin(data.(string))
		},
	)
}
