package main

import (
	"reflect"
)

type configuration struct {
	PingboardApiId     string `json:"pingboardApiClientID"`
	PingboardApiSecret string `json:"pingboardApiClientSecret"`
}

func (c *configuration) Clone() *configuration {
	var clone = *c
	return &clone
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

func (p *Plugin) setConfigurationIsChanged(configuration *configuration) bool {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return false
		}
		panic("setConfigurationIsChanged called with the existing configuration")
	}

	if configuration != nil && p.configuration != nil && *configuration == *p.configuration {
		return false
	}

	p.configuration = configuration
	return true
}
