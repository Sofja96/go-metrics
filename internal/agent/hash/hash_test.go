package hash

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComputeHmac256(t *testing.T) {
	key := "key"
	data := "data"
	want := "5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0"

	hash, err := ComputeHmac256([]byte(key), []byte(data))

	require.Equal(t, want, hash)
	require.NoError(t, err)
}
