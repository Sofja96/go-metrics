package agent

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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

	publicKey, err = LoadPublicKey("public_key.pem")
	assert.NoError(t, err)
	assert.NotNil(t, publicKey)

	_, err = LoadPublicKey("nonexistent_file.pem")
	assert.Error(t, err)

	publicKey, err = LoadPublicKey("")
	assert.NoError(t, err)
	assert.Nil(t, publicKey)
}
