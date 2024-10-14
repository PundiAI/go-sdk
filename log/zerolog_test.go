package log_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pundiai/go-sdk/log"
)

func TestZeroLogger_Console(t *testing.T) {
	output := new(bytes.Buffer)
	log.DefaultWriter = output

	logger, err := log.NewZeroLogger(log.FormatConsole, log.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Info("test")
	assert.Contains(t, output.String(), "test")
	t.Log(output.String())
}

func TestZeroLogger_JSON(t *testing.T) {
	output := new(bytes.Buffer)
	log.DefaultWriter = output

	logger, err := log.NewZeroLogger(log.FormatJSON, log.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Info("test")
	assert.Contains(t, output.String(), "test")
	t.Log(output.String())
}

func TestZeroLogger_Nil(t *testing.T) {
	logger, err := log.NewZeroLogger(log.FormatConsole, log.LevelInfo)
	assert.NoError(t, err)
	logger.Info("test", "error", nil)
}

func TestZeroLogger_With(t *testing.T) {
	output := new(bytes.Buffer)
	log.DefaultWriter = output

	logger, err := log.NewZeroLogger(log.FormatConsole, log.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger = logger.With("key", "value")
	logger.Info("test")
	assert.Contains(t, output.String(), "test")
	assert.Contains(t, output.String(), "key")
	assert.Contains(t, output.String(), "value")
	t.Log(output.String())
}
