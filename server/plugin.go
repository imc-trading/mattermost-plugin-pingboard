package main

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
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
	refreshLock       sync.RWMutex
	configuration     *configuration
	refreshTimer      *time.Timer
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

func (p *Plugin) UserHasBeenCreated(c *plugin.Context, user *model.User) {
	if c.UserAgent == "" {
		return
	}
	p.refreshData()
}

func (p *Plugin) OnActivate() error {
	p.refreshTimer = time.AfterFunc(time.Duration(5)*time.Second, p.refreshData)
	return nil
}
