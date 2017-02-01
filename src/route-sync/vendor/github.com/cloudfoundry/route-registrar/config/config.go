package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"

	"code.cloudfoundry.org/multierror"
)

type MessageBusServerSchema struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type HealthCheckSchema struct {
	Name       string `yaml:"name"`
	ScriptPath string `yaml:"script_path"`
	Timeout    string `yaml:"timeout"`
}

type ConfigSchema struct {
	MessageBusServers []MessageBusServerSchema `yaml:"message_bus_servers"`
	Routes            []RouteSchema            `yaml:"routes"`
	Host              string                   `yaml:"host"`
}

type RouteSchema struct {
	Name                 string             `yaml:"name"`
	Port                 *int               `yaml:"port"`
	Tags                 map[string]string  `yaml:"tags"`
	URIs                 []string           `yaml:"uris"`
	RouteServiceUrl      string             `yaml:"route_service_url"`
	RegistrationInterval string             `yaml:"registration_interval,omitempty"`
	HealthCheck          *HealthCheckSchema `yaml:"health_check,omitempty"`
}

type MessageBusServer struct {
	Host     string
	User     string
	Password string
}

type HealthCheck struct {
	Name       string
	ScriptPath string
	Timeout    time.Duration
}

type Config struct {
	MessageBusServers []MessageBusServer
	Routes            []Route
	Host              string
}

type Route struct {
	Name                 string
	Port                 int
	Tags                 map[string]string
	URIs                 []string
	RouteServiceUrl      string
	RegistrationInterval time.Duration
	HealthCheck          *HealthCheck
}

func NewConfigSchemaFromFile(configFile string) (ConfigSchema, error) {
	var config ConfigSchema

	c, err := ioutil.ReadFile(configFile)
	if err != nil {
		return ConfigSchema{}, err
	}

	err = yaml.Unmarshal(c, &config)
	if err != nil {
		return ConfigSchema{}, err
	}

	return config, nil
}

func (c ConfigSchema) ToConfig() (*Config, error) {
	errors := multierror.NewMultiError("config")

	if c.Host == "" {
		errors.Add(fmt.Errorf("host required"))
	}

	messageBusServers, err := messageBusServersFromSchema(c.MessageBusServers)
	if err != nil {
		errors.Add(err)
	}
	routes := []Route{}
	for index, r := range c.Routes {
		route, err := routeFromSchema(r, index)
		if err != nil {
			errors.Add(err)
			continue
		}

		routes = append(routes, *route)
	}

	if errors.Length() > 0 {
		return nil, errors
	}

	config := Config{
		Host:              c.Host,
		MessageBusServers: messageBusServers,
		Routes:            routes,
	}

	return &config, nil
}

func nameOrIndex(r RouteSchema, index int) string {
	if r.Name != "" {
		return fmt.Sprintf(`"%s"`, r.Name)
	}

	return strconv.Itoa(index)
}

func parseRegistrationInterval(registrationInterval string) (time.Duration, error) {
	var duration time.Duration

	if registrationInterval == "" {
		return duration, fmt.Errorf("no registration_interval")
	}

	var err error
	duration, err = time.ParseDuration(registrationInterval)
	if err != nil {
		return duration, fmt.Errorf("invalid registration_interval: %s", err.Error())
	}

	if duration <= 0 {
		return duration, fmt.Errorf("invalid registration_interval: interval must be greater than 0")
	}

	return duration, nil
}

func routeFromSchema(r RouteSchema, index int) (*Route, error) {
	errors := multierror.NewMultiError(fmt.Sprintf("route %s", nameOrIndex(r, index)))

	if r.Name == "" {
		errors.Add(fmt.Errorf("no name"))
	}

	if r.Port == nil {
		errors.Add(fmt.Errorf("no port"))
	} else if *r.Port <= 0 {
		errors.Add(fmt.Errorf("invalid port: %d", *r.Port))
	}

	if len(r.URIs) == 0 {
		errors.Add(fmt.Errorf("no URIs"))
	}

	for _, u := range r.URIs {
		if u == "" {
			errors.Add(fmt.Errorf("empty URIs"))
			break
		}
	}

	_, err := url.Parse(r.RouteServiceUrl)
	if err != nil {
		errors.Add(err)
	}

	registrationInterval, err := parseRegistrationInterval(r.RegistrationInterval)
	if err != nil {
		errors.Add(err)
	}

	var healthCheck *HealthCheck
	if r.HealthCheck != nil {
		healthCheck, err = healthCheckFromSchema(r.HealthCheck, registrationInterval)
		if err != nil {
			errors.Add(err)
		}
	}

	if errors.Length() > 0 {
		return nil, errors
	}

	route := Route{
		Name:                 r.Name,
		Port:                 *r.Port,
		Tags:                 r.Tags,
		URIs:                 r.URIs,
		RouteServiceUrl:      r.RouteServiceUrl,
		RegistrationInterval: registrationInterval,
		HealthCheck:          healthCheck,
	}
	return &route, nil
}

func healthCheckFromSchema(
	healthCheckSchema *HealthCheckSchema,
	registrationInterval time.Duration,
) (*HealthCheck, error) {
	errors := multierror.NewMultiError("healthcheck")

	healthCheck := &HealthCheck{
		Name:       healthCheckSchema.Name,
		ScriptPath: healthCheckSchema.ScriptPath,
	}

	if healthCheck.Name == "" {
		errors.Add(fmt.Errorf("no name"))
	}

	if healthCheck.ScriptPath == "" {
		errors.Add(fmt.Errorf("no script_path"))
	}

	if healthCheckSchema.Timeout == "" && registrationInterval > 0 {
		if errors.Length() > 0 {
			return nil, errors
		}

		healthCheck.Timeout = registrationInterval / 2
		return healthCheck, nil
	}

	var err error
	healthCheck.Timeout, err = time.ParseDuration(healthCheckSchema.Timeout)
	if err != nil {
		errors.Add(fmt.Errorf("invalid healthcheck timeout: %s", err.Error()))
		return nil, errors
	}

	if healthCheck.Timeout <= 0 {
		errors.Add(fmt.Errorf("invalid healthcheck timeout: %s", healthCheck.Timeout))
		return nil, errors
	}

	if healthCheck.Timeout >= registrationInterval && registrationInterval > 0 {
		errors.Add(fmt.Errorf(
			"invalid healthcheck timeout: %v must be less than the registration interval: %v",
			healthCheck.Timeout,
			registrationInterval,
		))
		return nil, errors
	}

	if errors.Length() > 0 {
		return nil, errors
	}

	return healthCheck, nil
}

func messageBusServersFromSchema(servers []MessageBusServerSchema) ([]MessageBusServer, error) {
	messageBusServers := []MessageBusServer{}
	if len(servers) < 1 {
		return nil, fmt.Errorf("message_bus_servers must have at least one entry")
	}

	for _, m := range servers {
		messageBusServers = append(
			messageBusServers,
			MessageBusServer{
				Host:     m.Host,
				User:     m.User,
				Password: m.Password,
			},
		)
	}

	return messageBusServers, nil
}
