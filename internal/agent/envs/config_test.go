package envs

import (
	"flag"
	"os"
	"testing"
)

func TestRunParameters(t *testing.T) {
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
				PollInterval:   10,
				HashKey:        "",
				RateLimit:      2,
			},
		},
	}

	for i, tc := range tests {
		for key, value := range tc.envVars {
			os.Setenv(key, value)
		}

		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = append([]string{"cmd"}, tc.args...)

		cfg := &Config{}
		err := RunParameters(cfg)

		if tc.expectErr {
			if err == nil {
				t.Errorf("test %d: expected error but got nil", i)
			}
			continue
		} else if err != nil {
			t.Errorf("test %d: unexpected error: %v", i, err)
			continue
		}

		if cfg.Address != tc.expected.Address {
			t.Errorf("test %d: expected Address to be '%s', got '%s'", i, tc.expected.Address, cfg.Address)
		}
		if cfg.ReportInterval != tc.expected.ReportInterval {
			t.Errorf("test %d: expected ReportInterval to be '%d', got '%d'", i, tc.expected.ReportInterval, cfg.ReportInterval)
		}
		if cfg.PollInterval != tc.expected.PollInterval {
			t.Errorf("test %d: expected PollInterval to be '%d', got '%d'", i, tc.expected.PollInterval, cfg.PollInterval)
		}
		if cfg.HashKey != tc.expected.HashKey {
			t.Errorf("test %d: expected HashKey to be '%s', got '%s'", i, tc.expected.HashKey, cfg.HashKey)
		}
		if cfg.RateLimit != tc.expected.RateLimit {
			t.Errorf("test %d: expected RateLimit to be '%d', got '%d'", i, tc.expected.RateLimit, cfg.RateLimit)
		}

		for key := range tc.envVars {
			os.Unsetenv(key)
		}
	}
}
