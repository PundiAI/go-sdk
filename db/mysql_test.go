package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMysql_CheckSource(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "test1",
			source:  "root:root@tcp(127.0.0.1:3306)/coastdao?charset=utf8mb4&parseTime=True&loc=Local",
			wantErr: assert.NoError,
		},
		{
			name:    "test2",
			source:  "",
			wantErr: assert.Error,
		},
		{
			name:    "test3",
			source:  "coastdao",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mysql{}
			tt.wantErr(t, m.ParseSource(tt.source), fmt.Sprintf("ParseSource(%v)", tt.source))
		})
	}
}
