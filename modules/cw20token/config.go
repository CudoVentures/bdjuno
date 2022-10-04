package cw20token

import (
	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectID string `yaml:"project_id"`
	SubID     string `yaml:"sub_id"`
}

func ParseConfig(bz []byte) (*Config, error) {
	cfg := struct {
		config *Config `yaml:"verified_contracts_subscription"`
	}{}

	if err := yaml.Unmarshal(bz, &cfg); err != nil {
		return nil, err
	}

	return cfg.config, nil
}
