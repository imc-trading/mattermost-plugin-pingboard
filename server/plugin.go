package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *configuration
	lastRefresh       time.Time
}

func (p *Plugin) getResult(response *resty.Response, err error, description string, result interface{}, validate func() bool) bool {
	if err != nil {
		p.API.LogError("Failed to obtain "+description, "error", err)
		return false
	}
	if response.StatusCode() != http.StatusOK {
		p.API.LogError("Failed to obtain "+description, "status", response.Status, "body", response)
		return false
	}
	err = json.Unmarshal(response.Body(), &result)
	if err != nil {
		p.API.LogError("Failed to decode response for "+description, "error", err, "body", response)
		return false
	}
	if !validate() {
		p.API.LogError("Failed to extract valid fields for "+description, "error", err, "body", response)
		return false
	}
	return true
}

func (p *Plugin) refreshData() {
	config := p.getConfiguration()
	if config.PingboardApiId == "" || config.PingboardApiSecret == "" {
		p.API.LogInfo("No Pingboard client configuration")
		return
	}

	p.API.LogInfo("Refreshing data...")

	client := resty.New().
		SetHeader("Content-Type", "application/json")

	type credentialsResponse struct {
		Token            string `json:"access_token"`
		SecondsRemaining int    `json:"expires_in"`
	}
	response, err := client.R().
		SetQueryParams(map[string]string{"grant_type": "client_credentials"}).
		SetBody(fmt.Sprintf("{\"client_id\": \"%s\", \"client_secret\": \"%s\"}", config.PingboardApiId, config.PingboardApiSecret)).
		Post("https://app.pingboard.com/oauth/token")
	tokenResult := credentialsResponse{}
	if !p.getResult(response, err, "token", &tokenResult, func() bool {
		return tokenResult.Token != "" && tokenResult.SecondsRemaining != 0
	}) {
		return
	}

	expires := time.Now().Add(time.Duration(tokenResult.SecondsRemaining) * time.Second)
	p.API.LogInfo("Got token", "expires", expires)

	client = client.SetAuthToken(tokenResult.Token)

	type companyResponse struct {
		Name   string `json:"name"`
		Domain string `json:"subdomain"`
	}
	type companiesResponse struct {
		Companies []companyResponse `json:"companies"`
	}
	companiesResult := companiesResponse{}
	response, err = client.R().
		Get("https://app.pingboard.com/api/v2/companies/my_company")
	if !p.getResult(response, err, "companies", &companiesResult, func() bool {
		return len(companiesResult.Companies) == 1
	}) {
		return
	}
	companyResult := companiesResult.Companies[0]
	p.API.LogInfo(fmt.Sprintf("Got company %s with sub-domain %s", companyResult.Name, companyResult.Domain))

	type usersMetaResponse struct {
		Page      int `json:"page"`
		PageCount int `json:"page_count"`
	}
	type metaResponse struct {
		Users usersMetaResponse `json:"users"`
	}
	type userResponse struct {
		// TODO get location also
		Id        string `json:"id"`
		StartDate string `json:"start_date"`
		Email     string `json:"email"`
		Phone     string `json:"office_phone"`
		JobTitle  string `json:"job_title"`
	}
	type usersResponse struct {
		Users []userResponse `json:"users"`
		Meta  metaResponse   `json:"meta"`
	}
	usersResult := usersResponse{}
	for page := 1; usersResult.Meta.Users.PageCount == 0 || page <= usersResult.Meta.Users.PageCount; page += 1 {
		response, err = client.R().
			SetQueryParams(map[string]string{"page_size": "200", "page": fmt.Sprintf("%d", page)}).
			Get("https://app.pingboard.com/api/v2/users")
		if !p.getResult(response, err, "users", &usersResult, func() bool {
			return usersResult.Meta.Users.Page == page && len(usersResult.Users) > 0
		}) {
			return
		}
		p.API.LogInfo(fmt.Sprintf("Got %d users on page %d", len(usersResult.Users), page))

		for _, user := range usersResult.Users {
			p.API.LogInfo(fmt.Sprintf("%s: id %s, started %s, phone %s, title %s",
				user.Email, user.Id, user.StartDate, user.Phone, user.JobTitle))
		}
	}

	p.lastRefresh = time.Now()
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
