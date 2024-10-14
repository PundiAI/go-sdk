//go:build zaplog

package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"github.com/pundiai/go-sdk/log"
)

func TestZapLogger_Console(t *testing.T) {
	output := new(zaptest.Buffer)
	log.DefaultWriter = output

	logger, err := log.NewZapLogger(log.FormatConsole, log.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Info("test")
	assert.Contains(t, output.String(), "test")
	t.Log(output.String())
}

func TestZapLogger_JSON(t *testing.T) {
	output := new(zaptest.Buffer)
	log.DefaultWriter = output

	logger, err := log.NewZapLogger(log.FormatJSON, log.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Info("test")
	assert.Contains(t, output.String(), "test")
	t.Log(output.String())
}

func TestZapLogger_Nil(t *testing.T) {
	logger, err := log.NewZapLogger(log.FormatConsole, log.LevelInfo)
	assert.NoError(t, err)
	logger.Info("test", "error", nil)
}

func TestZapLogger_With(t *testing.T) {
	output := new(zaptest.Buffer)
	log.DefaultWriter = output

	logger, err := log.NewZapLogger(log.FormatConsole, log.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger = logger.With("key", "value")
	logger.Info("test")
	assert.Contains(t, output.String(), "test")
	assert.Contains(t, output.String(), "key")
	assert.Contains(t, output.String(), "value")
	t.Log(output.String())
}
