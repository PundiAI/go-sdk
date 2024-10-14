package tool_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/pundiai/go-sdk/tool"
)

func TestError(t *testing.T) {
	assert.False(t, tool.IsIgnoreException(errors.New("Not Ignore Exception")))
	assert.False(t, tool.IsIgnoreException(errors.Wrap(fmt.Errorf("not Ignore Exception"), "Not Ignore Exception")))
	assert.True(t, tool.IsIgnoreException(tool.NewIgnoreException(errors.New("Ignore Exception"))))
	assert.True(t, tool.IsIgnoreException(tool.NewIgnoreException(fmt.Errorf("ignore Exception"))))
}
