package config

import (
	"errors"
	"mediator/src/client"
)

type Config struct {
	Host       string
	Port       int
	ClientList map[string]*client.Client `json:"Client"`
}

func (config *Config) GetClient(name string) (*client.Client, error) {
	for _, c := range config.ClientList {
		if c.Name == name {
			return c, nil
		}
	}
	return &client.Client{}, errors.New("not yet implemented")
}
