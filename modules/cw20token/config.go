package cw20token

import (
	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectID string `yaml:"project_id"`
	SubID     string `yaml:"sub_id"`
}

func ParseConfig(cfgBytes []byte) (*Config, error) {
	cfg := struct {
		Config *Config `yaml:"verified_contracts_subscription"`
	}{}

	err := yaml.Unmarshal(cfgBytes, &cfg)
	return cfg.Config, err
}
