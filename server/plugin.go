package main

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

type User struct {
	Id        string `json:"id"`
	Url       string `json:"url"`
	StartDate string `json:"start_date"`
	Phone     string `json:"phone"`
	JobTitle  string `json:"job_title"`
}

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *configuration
	usersLock         sync.RWMutex
	lastRefresh       time.Time
	usersByEmail      map[string]User
}

func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(configuration)
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	if p.setConfigurationIsChanged(configuration) {
		p.API.LogInfo("Config changed")
		p.refreshData()
	}

	return nil
}

func (p *Plugin) OnActivate() error {
	return nil
}
