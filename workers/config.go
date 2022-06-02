package workers

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type workerConfig struct {
	Name     string `yaml:"name"`
	Interval string `yaml:"interval"`
}

type workersConfig struct {
	Workers []workerConfig `yaml:"workers"`
}

func parseConfig(data []byte) (workersConfig, error) {
	var cfg workersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return workersConfig{}, fmt.Errorf("failed to unmarshal config: %s", err.Error())
	}
	return cfg, nil
}
