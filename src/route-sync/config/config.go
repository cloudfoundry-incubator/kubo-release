package config

import (
	"encoding/json"
	"fmt"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RawNatsServers         string                      `envconfig:"nats_servers" required:"true"`
	NatsServers            []cfConfig.MessageBusServer `ignore:"true"`
	CloudFoundryApiUrl     string                      `envconfig:"cloud_foundry_api_url" required:"true"`
	UAAApiUrl              string                      `envconfig:"uaa_api_url" required:"true"`
	RoutingApiUsername     string                      `envconfig:"routing_api_username" required:"true"`
	RoutingApiClientSecret string                      `envconfig:"routing_api_client_secret" required:"true"`
	SkipTlsVerification    bool                        `envconfig:"skip_tls_verification" required:"true"`
	KubeConfigPath         string                      `envconfig:"kube_config_path" required:"true"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	if err := envconfig.Process("routesync", &cfg); err != nil {
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
