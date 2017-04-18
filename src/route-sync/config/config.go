package config

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/multierror"
	cfConfig "code.cloudfoundry.org/route-registrar/config"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

// ConfigSchema defines the file schema for the YAML configuration of route-sync
type ConfigSchema struct {
	NatsServers               []MessageBusServerSchema `yaml:"nats_servers"`
	RoutingApiUrl             string                   `yaml:"routing_api_url"`
	CloudFoundryAppDomainName string                   `yaml:"app_domain_name"`
	UaaApiUrl                 string                   `yaml:"uaa_api_url"`
	RoutingApiUsername        string                   `yaml:"routing_api_username"`
	RoutingApiClientSecret    string                   `yaml:"routing_api_client_secret"`
	SkipTlsVerification       bool                     `yaml:"skip_tls_verification"`
	KubeConfigPath            string                   `yaml:"kube_config_path"`
}

// MessageBusServerSchema defines the schema for NATS Servers in the route-sync config
type MessageBusServerSchema struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// Config is the final application configuration for route-sync
type Config struct {
	NatsServers               []cfConfig.MessageBusServer
	RoutingApiUrl             string
	CloudFoundryAppDomainName string
	UaaApiUrl                 string
	RoutingApiUsername        string
	RoutingApiClientSecret    string
	SkipTlsVerification       bool
	KubeConfigPath            string
}

// NewConfigSchemaFromFile Loads a YAML file that contains the route-sync config
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

func missingOptionError(option, desc string) error {
	return fmt.Errorf("config option: %s, error: %s", option, desc)
}

// ToConfig Validates and builds a Config object if the ConfigSchema is valid
func (cs *ConfigSchema) ToConfig() (*Config, error) {
	errs := multierror.NewMultiError("config")

	messageBusServers := []cfConfig.MessageBusServer{}
	for _, messageBusServer := range cs.NatsServers {
		server, err := messageBusServer.ToConfig()
		if err != nil {
			errs.Add(err)
		} else {
			messageBusServers = append(messageBusServers, server)
		}
	}

	if len(cs.NatsServers) == 0 {
		errs.Add(missingOptionError("nats_servers", "at least 1 nats server is required"))
	}

	if len(cs.RoutingApiUrl) == 0 {
		errs.Add(missingOptionError("routing_api_url", "can not be blank"))
	}

	if len(cs.CloudFoundryAppDomainName) == 0 {
		errs.Add(missingOptionError("app_domain_name", "can not be blank"))
	}

	if len(cs.UaaApiUrl) == 0 {
		errs.Add(missingOptionError("uaa_api_url", "can not be blank"))
	}

	if len(cs.RoutingApiUsername) == 0 {
		errs.Add(missingOptionError("routing_api_username", "can not be blank"))
	}

	if len(cs.RoutingApiClientSecret) == 0 {
		errs.Add(missingOptionError("routing_api_client_secret", "can not be blank"))
	}

	if len(cs.KubeConfigPath) == 0 {
		errs.Add(missingOptionError("kube_config_path", "can not be blank"))
	}

	cfg := &Config{
		NatsServers:               messageBusServers,
		RoutingApiUrl:             cs.RoutingApiUrl,
		CloudFoundryAppDomainName: cs.CloudFoundryAppDomainName,
		UaaApiUrl:                 cs.UaaApiUrl,
		RoutingApiUsername:        cs.RoutingApiUsername,
		RoutingApiClientSecret:    cs.RoutingApiClientSecret,
		SkipTlsVerification:       cs.SkipTlsVerification,
		KubeConfigPath:            cs.KubeConfigPath,
	}

	if errs.Length() > 0 {
		return nil, errs
	}

	return cfg, nil
}

func (mbs *MessageBusServerSchema) ToConfig() (cfConfig.MessageBusServer, error) {
	errs := multierror.NewMultiError("config")

	if len(mbs.Host) == 0 {
		errs.Add(missingOptionError("nats_servers[].host", "can not be blank"))
	}

	if len(mbs.User) == 0 {
		errs.Add(missingOptionError("nats_servers[].user", "can not be blank"))
	}

	if len(mbs.Password) == 0 {
		errs.Add(missingOptionError("nats_servers[].password", "can not be blank"))
	}

	if errs.Length() > 0 {
		return cfConfig.MessageBusServer{}, errs
	}

	return cfConfig.MessageBusServer{Host: mbs.Host, User: mbs.User, Password: mbs.Password}, nil
}

func (cfg *Config) UAAConfig() *uaaconfig.Config {
	return &uaaconfig.Config{
		ClientName:       cfg.RoutingApiUsername,
		ClientSecret:     cfg.RoutingApiClientSecret,
		UaaEndpoint:      cfg.UaaApiUrl,
		SkipVerification: cfg.SkipTlsVerification,
	}
}
