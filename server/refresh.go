package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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
		p.API.LogError("Failed to decode response for "+description, "error", err)
		return false
	}
	if !validate() {
		p.API.LogError("Failed to extract valid fields for " + description)
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
	p.refreshTimer = time.AfterFunc(time.Duration(6)*time.Hour, p.refreshData)
	p.API.LogInfo("Refreshing data...")

	client := resty.New().
		SetHeader("Content-Type", "application/json")

	// get auth token using client credentials
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

	// get details of api user's company
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

	// get details of all users
	type usersMetaResponse struct {
		Page      int `json:"page"`
		PageCount int `json:"page_count"`
	}
	type userLinks struct {
		DepartmentIds []string `json:"departments"`
		LocationIds   []string `json:"locations"`
	}
	type metaResponse struct {
		Users usersMetaResponse `json:"users"`
	}
	type userResponse struct {
		Id        string    `json:"id"`
		StartDate string    `json:"start_date"`
		Email     string    `json:"email"`
		Phone     string    `json:"office_phone"`
		JobTitle  string    `json:"job_title"`
		Links     userLinks `json:"links"`
	}
	type usersResponse struct {
		Users []userResponse `json:"users"`
		Meta  metaResponse   `json:"meta"`
	}
	type groupResponse struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type groupsResponse struct {
		Groups []groupResponse `json:"groups"`
	}
	usersByEmail := map[string]User{}
	departmentsById := map[string]string{}
	usersResult := usersResponse{}
	dateExpr := regexp.MustCompile(`([0-9]{4})-([0-9]{2})-([0-9]{2})`)
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
			department := ""
			if user.Links.DepartmentIds != nil && len(user.Links.DepartmentIds) >= 1 {
				departmentId := user.Links.DepartmentIds[0]
				firstDepartment, found := departmentsById[departmentId]
				if !found {
					response, err = client.R().
						Get(fmt.Sprintf("https://app.pingboard.com/api/v2/groups/%s", departmentId))
					departmentResult := groupsResponse{}
					if !p.pingboardResponse(response, err, "department", &departmentResult, func() bool {
						return len(departmentResult.Groups) == 1 && departmentResult.Groups[0].Id == departmentId
					}) {
						return
					}
					p.API.LogDebug(fmt.Sprintf("department %s: %s",
						departmentId, departmentResult.Groups[0].Name))
					firstDepartment = departmentResult.Groups[0].Name
					departmentsById[departmentId] = firstDepartment
				}
				department = firstDepartment
			}
			dateParts := dateExpr.FindStringSubmatch(user.StartDate)
			if dateParts == nil {
				p.API.LogError("Failed to parse date: " + user.StartDate)
				return
			}
			startYear, _ := strconv.Atoi(dateParts[1])
			startMonth, _ := strconv.Atoi(dateParts[2])
			startDay, _ := strconv.Atoi(dateParts[3])
			usersByEmail[strings.ToLower(user.Email)] = User{
				Id:         user.Id,
				Url:        fmt.Sprintf("https://%s.pingboard.com/users/%s", companyResult.Domain, user.Id),
				StartYear:  startYear,
				StartMonth: startMonth,
				StartDay:   startDay,
				Phone:      user.Phone,
				JobTitle:   user.JobTitle,
				Department: department,
			}
		}
	}

	p.usersByEmail = usersByEmail
}
