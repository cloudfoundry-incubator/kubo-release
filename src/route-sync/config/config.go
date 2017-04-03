package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

type ConfigSchema struct {
	NatsServers               []MessageBusServerSchema `yaml:"nats_servers"`
	RoutingApiUrl             string                   `yaml:"routing_api_url"`
	CloudFoundryAppDomainName string                   `yaml:"app_domain_name"`
	UAAApiURL                 string                   `yaml:"uaa_api_url"`
	RoutingAPIUsername        string                   `yaml:"routing_api_username"`
	RoutingAPIClientSecret    string                   `yaml:"routing_api_client_secret"`
	SkipTLSVerification       bool                     `yaml:"skip_tls_verification"`
	KubeConfigPath            string                   `yaml:"kube_config_path"`
}

type MessageBusServerSchema struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Config struct {
	NatsServers               []cfConfig.MessageBusServer
	RoutingApiUrl             string
	CloudFoundryAppDomainName string
	UAAApiURL                 string
	RoutingAPIUsername        string
	RoutingAPIClientSecret    string
	SkipTLSVerification       bool
	KubeConfigPath            string
}

func NewConfigSchemaFromFile(path string) (*ConfigSchema, error) {
	var schema ConfigSchema

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(raw, &schema)
	if err != nil {
		return nil, err
	}

	return &schema, err
}

func (cs *ConfigSchema) ToConfig() (*Config, error) {
	messageBusServers := []cfConfig.MessageBusServer{}
	for _, messageBusServer := range cs.NatsServers {
		messageBusServers = append(messageBusServers, messageBusServer.ToConfig())
	}

	cfg := &Config{
		NatsServers:               messageBusServers,
		RoutingApiUrl:             cs.RoutingApiUrl,
		CloudFoundryAppDomainName: cs.CloudFoundryAppDomainName,
		UAAApiURL:                 cs.UAAApiURL,
		RoutingAPIUsername:        cs.RoutingAPIUsername,
		RoutingAPIClientSecret:    cs.RoutingAPIClientSecret,
		SkipTLSVerification:       cs.SkipTLSVerification,
		KubeConfigPath:            cs.KubeConfigPath,
	}

	return cfg, nil
}

func (mbs *MessageBusServerSchema) ToConfig() cfConfig.MessageBusServer {
	return cfConfig.MessageBusServer{Host: mbs.Host, User: mbs.User, Password: mbs.Password}
}

func (cfg *Config) UAAConfig() *uaaconfig.Config {
	return &uaaconfig.Config{
		ClientName:       cfg.RoutingAPIUsername,
		ClientSecret:     cfg.RoutingAPIClientSecret,
		UaaEndpoint:      cfg.UAAApiURL,
		SkipVerification: cfg.SkipTLSVerification,
	}
}
