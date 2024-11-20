package modules

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

var _ CustomType = (*BigInt)(nil)

type BigInt big.Int

func (b *BigInt) Scan(value any) error {
	str, err := unquoteIfQuoted(value)
	if err != nil {
		return err
	}
	result, ok := new(big.Int).SetString(str, 10)
	if !ok {
		return fmt.Errorf("big.Int set string error, value: %s", str)
	}
	*b = BigInt(*result)
	return nil
}

func (b *BigInt) Value() (driver.Value, error) {
	result := big.Int(*b)
	return result.String(), nil
}

func (b *BigInt) String() string {
	return (*big.Int)(b).String()
}

func (b *BigInt) MustToBigInt() *big.Int {
	return (*big.Int)(b)
}

func NewBigInt(value *big.Int) *BigInt {
	return (*BigInt)(value)
}

func unquoteIfQuoted(value any) (string, error) {
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return "", fmt.Errorf("could not convert value '%+v' to byte array of type '%T'", value, value)
	}
	// If the amount is quoted, strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
