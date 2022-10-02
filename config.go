package main

import (
	"fmt"

	"github.com/akerl/go-lambda/s3"
)

type check struct {
	Name      string `json:"name"`
	CustomKey string `json:"key"`
	Threshold int    `json:"threshold"`
}

func (c check) Key() string {
	if c.CustomKey != "" {
		return c.CustomKey
	}
	return c.Name
}

type configFile struct {
	Checks       []check `json:"checks"`
	SlackWebhook string  `json:"slack_webhook"`
}

var config *configFile

func loadConfig() (*configFile, error) {
	c := configFile{}
	cf, err := s3.GetConfigFromEnv(&c)
	if err != nil {
		return &c, err
	}
	cf.OnError = func(_ *s3.ConfigFile, err error) {
		fmt.Println(err)
	}
	cf.Autoreload(60)

	for _, ch := range c.Checks {
		if ch.Name == "" {
			return &c, fmt.Errorf("checks must have a name")
		}
	}

	return &c, nil
}
