package main

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

type User struct {
	Id         string `json:"id"`
	Email      string `json:"email"` // the email address exactly as Pingboard had it
	Url        string `json:"url"`
	StartYear  int    `json:"start_year"`
	StartMonth int    `json:"start_month"`
	StartDay   int    `json:"start_day"`
	Phone      string `json:"phone"`
	JobTitle   string `json:"job_title"`
	Department string `json:"department"`
	Manager    string `json:"manager"`
}

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	refreshLock       sync.RWMutex
	configuration     *configuration
	refreshTimer      *time.Timer
	usersByUsername   map[string]User
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
