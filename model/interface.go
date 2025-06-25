package model

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type CustomType interface {
	sql.Scanner
	driver.Valuer
	fmt.Stringer
}
