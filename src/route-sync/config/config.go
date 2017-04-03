package config

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/multierror"
	cfConfig "code.cloudfoundry.org/route-registrar/config"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

type ConfigSchema struct {
	NatsServers               []MessageBusServerSchema `yaml:"nats_servers"`
	RoutingAPIURL             string                   `yaml:"routing_api_url"`
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
	RoutingAPIURL             string
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

func missingOptionError(option, desc string) error {
	return fmt.Errorf("config option: %s, error: %s", option, desc)
}

func (cs *ConfigSchema) ToConfig() (*Config, error) {
	errs := multierror.NewMultiError("config")

	messageBusServers := []cfConfig.MessageBusServer{}
	for _, messageBusServer := range cs.NatsServers {
		messageBusServers = append(messageBusServers, messageBusServer.ToConfig())
	}

	if len(cs.NatsServers) == 0 {
		errs.Add(missingOptionError("nats_servers", "at least 1 nats server is required"))
	}

	if len(cs.RoutingAPIURL) == 0 {
		errs.Add(missingOptionError("routing_api_url", "can not be blank"))
	}

	if len(cs.CloudFoundryAppDomainName) == 0 {
		errs.Add(missingOptionError("app_domain_name", "can not be blank"))
	}

	if len(cs.UAAApiURL) == 0 {
		errs.Add(missingOptionError("uaa_api_url", "can not be blank"))
	}

	if len(cs.RoutingAPIUsername) == 0 {
		errs.Add(missingOptionError("routing_api_username", "can not be blank"))
	}

	if len(cs.RoutingAPIClientSecret) == 0 {
		errs.Add(missingOptionError("routing_api_client_secret", "can not be blank"))
	}

	if len(cs.KubeConfigPath) == 0 {
		errs.Add(missingOptionError("kube_config_path", "can not be blank"))
	}

	cfg := &Config{
		NatsServers:               messageBusServers,
		RoutingAPIURL:             cs.RoutingAPIURL,
		CloudFoundryAppDomainName: cs.CloudFoundryAppDomainName,
		UAAApiURL:                 cs.UAAApiURL,
		RoutingAPIUsername:        cs.RoutingAPIUsername,
		RoutingAPIClientSecret:    cs.RoutingAPIClientSecret,
		SkipTLSVerification:       cs.SkipTLSVerification,
		KubeConfigPath:            cs.KubeConfigPath,
	}

	if errs.Length() > 0 {
		return nil, errs
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
