package config

import (
	"errors"
	"os"

	"go.yaml.in/yaml/v2"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Routes []Route      `yaml:"routes"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type Route struct {
	Path         string   `yaml:"path"`
	Backends     []string `yaml:"backends"`
	TrimPrefix   bool     `yaml:"trim_prefix"`
	Protected    bool     `yaml:"protected"`
	AllowedRoles []string `yaml:"allowed_roles"`
}

func Load(path string) (*Config, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {

	if len(c.Routes) == 0 {
		return errors.New("no routes configured")
	}

	for _, route := range c.Routes {

		if route.Path == "" {
			return errors.New("route path required")
		}

		if len(route.Backends) == 0 {
			return errors.New("route must have at least one backend")
		}
	}

	return nil
}
