package tool

import (
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func ComposeDecodeHookFunc(fs ...mapstructure.DecodeHookFunc) mapstructure.DecodeHookFunc {
	fs = append(fs,
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		StringToBigIntHookFunc(),
		StringToAddressHookFunc(),
		StringToDecimalHookFunc(),
	)
	return mapstructure.ComposeDecodeHookFunc(fs...)
}

func StringToBigIntHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(big.Int{}) {
			return data, nil
		}
		i, ok := new(big.Int).SetString(data.(string), 10)
		if !ok {
			return nil, errors.Errorf("can't convert %v to big.Int", data)
		}
		return i, nil
	}
}

func StringToAddressHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(common.Address{}) {
			return data, nil
		}
		return common.HexToAddress(data.(string)), nil
	}
}

func StringToDecimalHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(decimal.Decimal{}) {
			return data, nil
		}
		switch f.Kind() {
		case reflect.Float64:
			return decimal.NewFromFloat(data.(float64)), nil
		case reflect.String:
			dec, err := decimal.NewFromString(data.(string))
			if err != nil {
				return nil, errors.Wrapf(err, "can't convert %v to decimal.Decimal", data)
			}
			return dec, err
		default:
			return nil, errors.Errorf("can't convert %v to decimal.Decimal", data)
		}
	}
}
