package config

import (
	"encoding/json"
	"fmt"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	"github.com/kelseyhightower/envconfig"
)

// Config contains settings for the route-sync with envconfig annotations
type Config struct {
	RawNatsServers         string                      `envconfig:"nats_servers" required:"true"`
	NatsServers            []cfConfig.MessageBusServer `ignore:"true"`
	CloudFoundryAPIURL     string                      `envconfig:"cloud_foundry_api_url" required:"true"`
	CloudFoundryAPPURL     string                      `envconfig:"cloud_foundry_app_url" required:"true"`
	UAAApiURL              string                      `envconfig:"uaa_api_url" required:"true"`
	RoutingAPIUsername     string                      `envconfig:"routing_api_username" required:"true"`
	RoutingAPIClientSecret string                      `envconfig:"routing_api_client_secret" required:"true"`
	SkipTLSVerification    bool                        `envconfig:"skip_tls_verification" required:"true"`
	KubeConfigPath         string                      `envconfig:"kube_config_path" required:"true"`
}

// NewConfig creates a Config object from the systems enviornment variables
//
// Pass in the values through enviornment variables with the ROUTESYNC_ prefix.
// Eg: ROUTESYNC_CLOUD_FOUNDRY_API_URL="http://api.cf.example.org"
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
