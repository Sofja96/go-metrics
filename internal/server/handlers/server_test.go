package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Sofja96/go-metrics.git/internal/server/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
	}

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err, "Failed to create logger")

	defer logger.Sync()

	server := New(context.Background())

	assert.NotNil(t, server)
	assert.Equal(t, cfg.Address, server.address)
	assert.NotNil(t, server.echo)

}

func TestStart(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
	}

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	defer logger.Sync()

	server := &APIServer{
		echo:    echo.New(),
		address: cfg.Address,
		logger:  *logger.Sugar(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	cancel()
}
