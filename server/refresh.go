package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func (p *Plugin) pingboardResponse(response *resty.Response, err error, description string, result interface{}, validate func() bool) bool {
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
	p.refreshLock.Lock()
	defer p.refreshLock.Unlock()

	if p.refreshTimer == nil {
		return
	}
	p.refreshTimer.Stop()

	config := p.getConfiguration()
	if config.PingboardApiId == "" || config.PingboardApiSecret == "" {
		p.API.LogInfo("No Pingboard client configuration")
		// do not schedule more attempts (config change will already trigger a refresh)
		return
	}

	// always schedule a later attempt even if we fail with errors below
	p.refreshTimer = time.AfterFunc(time.Duration(6) * time.Hour, p.refreshData)
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
	if !p.pingboardResponse(response, err, "token", &tokenResult, func() bool {
		return tokenResult.Token != "" && tokenResult.SecondsRemaining != 0
	}) {
		return
	}

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
	if !p.pingboardResponse(response, err, "companies", &companiesResult, func() bool {
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
	usersByEmail := map[string]User{}
	for page := 1; usersResult.Meta.Users.PageCount == 0 || page <= usersResult.Meta.Users.PageCount; page += 1 {
		response, err = client.R().
			SetQueryParams(map[string]string{"page_size": "200", "page": fmt.Sprintf("%d", page)}).
			Get("https://app.pingboard.com/api/v2/users")
		if !p.pingboardResponse(response, err, "users", &usersResult, func() bool {
			return usersResult.Meta.Users.Page == page && len(usersResult.Users) > 0
		}) {
			return
		}
		p.API.LogInfo(fmt.Sprintf("Got %d users on page %d", len(usersResult.Users), page))
		for _, user := range usersResult.Users {
			p.API.LogDebug(fmt.Sprintf("%s: id %s, started %s, phone %s, title %s",
				user.Email, user.Id, user.StartDate, user.Phone, user.JobTitle))
			usersByEmail[strings.ToLower(user.Email)] = User{
				Id:        user.Id,
				Url:       fmt.Sprintf("https://%s.pingboard.com/users/%s", companyResult.Domain, user.Id),
				StartDate: user.StartDate,
				Phone:     user.Phone,
				JobTitle:  user.JobTitle,
			}
		}
	}

	p.usersByEmail = usersByEmail
}
