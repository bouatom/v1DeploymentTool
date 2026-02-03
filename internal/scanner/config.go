package scanner

import "time"

type ScannerConfig struct {
	MaxConcurrency int
	RatePerSecond  int
	Timeout        time.Duration
}

func NormalizeConfig(config ScannerConfig) ScannerConfig {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 10
	}
	if config.RatePerSecond <= 0 {
		config.RatePerSecond = 50
	}
	if config.Timeout <= 0 {
		config.Timeout = 3 * time.Second
	}

	return config
}
