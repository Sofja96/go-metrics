package envs

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	type test struct {
		name      string
		envVars   map[string]string
		args      []string
		expected  Config
		expectErr bool
	}

	tests := []test{
		{
			name: "LoadEnvsSuccess",
			envVars: map[string]string{
				"ADDRESS":         "localhost:9000",
				"REPORT_INTERVAL": "15",
				"POLL_INTERVAL":   "5",
				"KEY":             "test-key",
				"RATE_LIMIT":      "50",
			},
			args: []string{},
			expected: Config{
				Address:        "localhost:9000",
				ReportInterval: 15,
				PollInterval:   5,
				HashKey:        "test-key",
				RateLimit:      50,
			},
		},
		{
			name:    "LoadFlagsSuccess",
			envVars: map[string]string{},
			args: []string{
				"-a", "localhost:7070",
				"-r", "20",
				"-p", "10",
				"-k", "another-key",
				"-l", "25",
			},
			expected: Config{
				Address:        "localhost:7070",
				ReportInterval: 20,
				PollInterval:   10,
				HashKey:        "another-key",
				RateLimit:      25,
			},
		},
		{
			name: "LoadFlagsWithEnvSuccess",
			envVars: map[string]string{
				"ADDRESS": "localhost:6060",
			},
			args: []string{
				"-r", "30",
			},
			expected: Config{
				Address:        "localhost:6060",
				ReportInterval: 30,
				PollInterval:   2,
				HashKey:        "",
				RateLimit:      1,
			},
		},
		{
			name: "LoadFileConfigSuccess",
			envVars: map[string]string{
				"CONFIG": "./mocks/config_test.json",
			},
			args: []string{},
			expected: Config{
				Address:        "localhost:8081",
				ReportInterval: 1,
				PollInterval:   1,
				CryptoKey:      "../../public.key",
				RateLimit:      1,
			},
		},
		{
			name: "LoadFileConfigWithFlagsEndEnvsSuccess",
			envVars: map[string]string{
				"CONFIG":          "./mocks/config_test.json",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "15",
			},
			args: []string{
				"-r", "30",
			},
			expected: Config{
				Address:        "localhost:8081",
				ReportInterval: 15,
				PollInterval:   5,
				CryptoKey:      "../../public.key",
				RateLimit:      1,
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
			assert.Equal(t, cfg.ReportInterval, tc.expected.ReportInterval, "expected ReportInterval to be '%d', got '%d'", tc.expected.ReportInterval, cfg.ReportInterval)
			assert.Equal(t, cfg.PollInterval, tc.expected.PollInterval, " expected PollInterval to be '%d', got '%d'", tc.expected.PollInterval, cfg.PollInterval)
			assert.Equal(t, cfg.HashKey, tc.expected.HashKey, "expected HashKey to be '%s', got '%s'", tc.expected.HashKey, cfg.HashKey)
			assert.Equal(t, cfg.RateLimit, tc.expected.RateLimit, "expected RateLimit to be '%d', got '%d'", tc.expected.RateLimit, cfg.RateLimit)

			for key := range tc.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
