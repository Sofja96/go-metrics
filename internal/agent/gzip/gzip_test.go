package gzip

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

var compressedData = []byte{
	0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x4, 0xff, 0x4, 0xc0, 0x41, 0xa, 0xc2, 0x30,
	0x10, 0x85, 0xe1, 0xbb, 0xfc, 0xeb, 0x41, 0x6c,
	0x75, 0x35, 0x57, 0x11, 0x17, 0x21, 0x79, 0x94,
	0x81, 0x56, 0x4b, 0x98, 0x8, 0xa5, 0xf4, 0xee,
	0x7e, 0xaf, 0x93, 0x68, 0x38, 0x9b, 0xb2, 0x47,
	0x9d, 0x30, 0xf2, 0xd8, 0x85, 0xb3, 0x94, 0xb1,
	0x8, 0xe3, 0x57, 0xd6, 0x21, 0x7c, 0xba, 0xdf,
	0xe6, 0xe7, 0x65, 0x27, 0xd1, 0x70, 0x36, 0x65,
	0x8f, 0x3a, 0x63, 0xe4, 0xb1, 0xb, 0xa7, 0x7e,
	0xc7, 0x27, 0xd5, 0x31, 0x9a, 0xd6, 0x2c, 0xf8,
	0xe3, 0x7a, 0xff, 0x3, 0x0, 0x0, 0xff, 0xff,
	0xc7, 0xee, 0x4, 0x8f, 0x5b, 0x0, 0x0, 0x0,
}

// Тестируем функцию Compress
func TestCompress(t *testing.T) {
	tests := []struct {
		name    string
		metrics []models.Metrics
	}{
		{
			name: "valid metrics",
			metrics: []models.Metrics{
				{MType: "gauge", ID: "metric1", Value: utils.FloatPtr(10.24)},
				{MType: "counter", ID: "metric2", Delta: utils.IntPtr(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compress(tt.metrics)
			require.NoError(t, err)

			require.Equal(t, compressedData, got)

		})
	}
}
