package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	type test struct {
		name     string
		envVars  map[string]string
		args     []string
		expected Config
	}

	tests := []test{
		{
			name: "LoadEnvsSuccess",
			envVars: map[string]string{
				"ADDRESS":           "localhost:9000",
				"STORE_INTERVAL":    "15",
				"FILE_STORAGE_PATH": "/tmp/metrics.json",
				"RESTORE":           "false",
				"KEY":               "test-key",
				"CRYPTO_KEY":        "",
				"TRUSTED_SUBNET":    "127.0.0.0/8",
			},
			args: []string{},
			expected: Config{
				Address:       "localhost:9000",
				StoreInterval: 15,
				Restore:       false,
				FilePath:      "/tmp/metrics.json",
				HashKey:       "test-key",
				DatabaseDSN:   "",
				CryptoKey:     "",
				TrustedSubnet: "127.0.0.0/8",
			},
		},
		{
			name:    "LoadFlagsSuccess",
			envVars: map[string]string{},
			args: []string{
				"-a", "localhost:7070",
				"-f", "/tmp/metrics.json",
				"-i", "10",
				"-k", "another-key",
				"-r=false",
				"-t", "127.0.0.0/8",
			},
			expected: Config{
				Address:       "localhost:7070",
				StoreInterval: 10,
				HashKey:       "another-key",
				Restore:       false,
				FilePath:      "/tmp/metrics.json",
				TrustedSubnet: "127.0.0.0/8",
			},
		},
		{
			name: "LoadFlagsWithEnvSuccess",
			envVars: map[string]string{
				"ADDRESS": "localhost:6060",
			},
			args: []string{
				"-i", "30",
			},
			expected: Config{
				Address:       "localhost:6060",
				StoreInterval: 30,
				HashKey:       "",
				FilePath:      "/tmp/metrics-db.json",
				Restore:       true,
			},
		},
		{
			name: "LoadFileConfigSuccess",
			envVars: map[string]string{
				"CONFIG": "./mocks/config_test.json",
			},
			args: []string{},
			expected: Config{
				Address:       "localhost:8081",
				Restore:       true,
				StoreInterval: 1,
				FilePath:      "/tmp/metrics-db.json",
				DatabaseDSN:   "",
				CryptoKey:     "../../private.key",
				TrustedSubnet: "127.0.0.0/8",
			},
		},
		{
			name: "LoadFileConfigWithFlagsEndEnvsSuccess",
			envVars: map[string]string{
				"CONFIG":         "./mocks/config_test.json",
				"STORE_INTERVAL": "5",
				"ADDRESS":        "localhost:9000",
			},
			args: []string{
				"-a", "localhost:7070",
			},
			expected: Config{
				Address:       "localhost:9000",
				StoreInterval: 5,
				Restore:       true,
				CryptoKey:     "../../private.key",
				FilePath:      "/tmp/metrics-db.json",
				DatabaseDSN:   "",
				TrustedSubnet: "127.0.0.0/8",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			for key, value := range tc.envVars {
				os.Setenv(key, value)
			}

			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = append([]string{"cmd"}, tc.args...)

			cfg, err := LoadConfig()
			assert.NoError(t, err)

			assert.Equal(t, cfg.Address, tc.expected.Address, "expected Address to be '%s', got '%s'", tc.expected.Address, cfg.Address)
			assert.Equal(t, cfg.StoreInterval, tc.expected.StoreInterval, "expected StoreInterval to be '%d', got '%d'", tc.expected.StoreInterval, cfg.StoreInterval)
			assert.Equal(t, cfg.FilePath, tc.expected.FilePath, " expected FilePath to be '%s', got '%s'", tc.expected.FilePath, cfg.FilePath)
			assert.Equal(t, cfg.HashKey, tc.expected.HashKey, "expected HashKey to be '%s', got '%s'", tc.expected.HashKey, cfg.HashKey)
			assert.Equal(t, cfg.Restore, tc.expected.Restore, "expected Restore to be '%t', got '%t'", tc.expected.Restore, cfg.Restore)
			assert.Equal(t, cfg.DatabaseDSN, tc.expected.DatabaseDSN, "expected DatabaseDSN to be '%s', got '%s'", tc.expected.DatabaseDSN, cfg.DatabaseDSN)
			assert.Equal(t, cfg.CryptoKey, tc.expected.CryptoKey, "expected CryptoKey to be '%s', got '%s'", tc.expected.CryptoKey, cfg.CryptoKey)
			assert.Equal(t, cfg.TrustedSubnet, tc.expected.TrustedSubnet, "expected TrustedSubnet to be '%s', got '%s'", tc.expected.TrustedSubnet, cfg.TrustedSubnet)

			for key := range tc.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
