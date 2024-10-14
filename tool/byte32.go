package tool

import (
	"fmt"
	"math/big"
)

func Byte32ToString(bytes [32]byte) string {
	for i := len(bytes) - 1; i >= 0; i-- {
		if bytes[i] != 0 {
			return string(bytes[:i+1])
		}
	}
	return ""
}

func StrToByte32(s string) ([32]byte, error) {
	var out [32]byte
	if len([]byte(s)) > 32 {
		return out, fmt.Errorf("string too long")
	}
	copy(out[:], s)
	return out, nil
}

func DivideToFloat(x, y *big.Int) float64 {
	quotient := new(big.Float).Quo(new(big.Float).SetInt(x), new(big.Float).SetInt(y))
	result, _ := quotient.Float64()
	return result
}

func Byte32ToBigInt(bytes [32]byte) *big.Int {
	return new(big.Int).SetBytes(bytes[:])
}
