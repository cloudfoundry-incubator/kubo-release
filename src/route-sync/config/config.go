package config

import (
	"encoding/json"
	"fmt"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RawNatsServers string                      `envconfig:"nats_servers" required:"true"`
	NatsServers    []cfConfig.MessageBusServer `ignore:"true"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(cfg.RawNatsServers), &cfg.NatsServers); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) validate() error {
	if len(cfg.NatsServers) == 0 {
		return fmt.Errorf("no NatsServers specified in config")
	}
	return nil
}
