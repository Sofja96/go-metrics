package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

func TestLoadPublicKey(t *testing.T) {
	privatePath := "private_key.pem"
	publicPath := "public_key.pem"

	defer os.Remove(privatePath)
	defer os.Remove(publicPath)

	privateKey, publicKey := utils.GenerateRsaKeyPair()
	privatePEM := utils.PrivateToString(privateKey)
	publicPEM, _ := utils.PublicToString(publicKey)

	err := utils.ExportToFile(privatePEM, privatePath)
	assert.NoError(t, err)

	err = utils.ExportToFile(publicPEM, publicPath)
	assert.NoError(t, err)

	t.Run("ValidPublicKey", func(t *testing.T) {
		publicKey, err := LoadPublicKey(publicPath)
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
	})

	t.Run("FileDoesNotExist", func(t *testing.T) {
		_, err := LoadPublicKey("nonexistent_file.pem")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error reading public key file")
	})

	t.Run("EmptyPath", func(t *testing.T) {
		publicKey, err := LoadPublicKey("")
		assert.NoError(t, err)
		assert.Nil(t, publicKey)
	})

	t.Run("InvalidPemFormat", func(t *testing.T) {
		err := utils.ExportToFile("invalid pem data", publicPath)
		assert.NoError(t, err)

		_, err = LoadPublicKey(publicPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid PEM format or missing public key")
	})
}

func TestGetMetrics(t *testing.T) {
	collector := metrics.NewMetricsCollector()
	ch := make(chan []byte, 1)

	getMetrics(collector, ch)

	select {
	case data := <-ch:
		assert.NotEmpty(t, data, "Данные не должны быть пустыми")
	default:
		t.Fatal("Ожидались данные в канале")
	}
}

func TestRun(t *testing.T) {
	go func() {
		err := Run()
		assert.NoError(t, err, "Ошибка при запуске агента")
	}()

	time.Sleep(100 * time.Millisecond)

}

func TestStartTask(t *testing.T) {
	t.Run("Process channel data", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := make(chan []byte, 1)
		ch <- []byte("test data")
		close(ch)

		go startTask(ctx, ch)

		select {
		case <-ch:
		default:
			t.Fatal("Канал не был обработан")
		}
	})

	t.Run("Cancel context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := make(chan []byte, 1)
		go startTask(ctx, ch)

		cancel()
		time.Sleep(100 * time.Millisecond)

		select {
		case <-ch:
			t.Fatal("Канал не должен иметь данных после отмены контекста")
		default:
		}
	})
}
