package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterDriver(t *testing.T) {
	assert.Len(t, drivers, 2)
}

func TestGetDriver(t *testing.T) {
	tests := []struct {
		name    string
		driver  string
		want    Driver
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "test1",
			driver:  SqliteDriver,
			want:    &Sqlite{},
			wantErr: assert.NoError,
		},
		{
			name:    "test2",
			driver:  MysqlDriver,
			want:    &Mysql{},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDriver(tt.driver)
			if !tt.wantErr(t, err, fmt.Sprintf("GetDriver(%v)", tt.driver)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetDriver(%v)", tt.driver)
		})
	}
}
