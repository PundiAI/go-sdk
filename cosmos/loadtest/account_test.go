package loadtest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateGenesisAccounts(t *testing.T) {
	homeDir := t.TempDir()
	defer t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(homeDir))
	})
	err := CreateGenesisAccounts("sei", 10, homeDir)
	assert.NoError(t, err)
}
