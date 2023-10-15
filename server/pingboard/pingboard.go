package pingboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Public types returned by this client
type User struct {
	Id          string
	StartDate   string
	Email       string
	Phone       string
	JobTitle    string
	ReportsToId string
	Department  string
}

type Company struct {
	Name   string `json:"name"`
	Domain string `json:"subdomain"`
}

// Internal Pingboard API response types for unmarshalling

type credentialsResponse struct {
	Token            string `json:"access_token"`
	SecondsRemaining int    `json:"expires_in"`
}
type companyResponse struct {
	Name   string `json:"name"`
	Domain string `json:"subdomain"`
}
type companiesResponse struct {
	Companies []companyResponse `json:"companies"`
}
type groupResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type groupsResponse struct {
	Groups []groupResponse `json:"groups"`
}
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
	Id          string    `json:"id"`
	StartDate   string    `json:"start_date"`
	Email       string    `json:"email"`
	Phone       string    `json:"office_phone"`
	JobTitle    string    `json:"job_title"`
	ReportsToId int       `json:"reports_to_id"`
	Links       userLinks `json:"links"`
}
type usersResponse struct {
	Users []userResponse `json:"users"`
	Meta  metaResponse   `json:"meta"`
}

type Client struct {
	restClient *resty.Client
	pluginAPI  plugin.API
}

func NewClient(pluginAPI plugin.API, pingboardId string, pingboardSecret string) *Client {
	client := Client{
		restClient: resty.New().
			SetHeader("Content-Type", "application/json"),
		pluginAPI: pluginAPI,
	}

	if !client.setAuthToken(pingboardId, pingboardSecret) {
		return nil
	}

	return &client
}

func (c *Client) pingboardResponse(response *resty.Response, err error, description string, result interface{}, validate func() bool) bool {
	if err != nil {
		c.pluginAPI.LogError(fmt.Sprintf("Failed to obtain %s", description),
			"error", err)
		return false
	}
	if response.StatusCode() != http.StatusOK {
		c.pluginAPI.LogError(fmt.Sprintf("Failed to obtain %s", description),
			"status", response.Status, "body", response)
		return false
	}
	err = json.Unmarshal(response.Body(), &result)
	if err != nil {
		c.pluginAPI.LogError(fmt.Sprintf("Failed to decode response for %s", description),
			"error", err)
		return false
	}
	if !validate() {
		c.pluginAPI.LogError(fmt.Sprintf("Failed to extract valid fields for %s", description))
		return false
	}
	return true
}

func (c *Client) setAuthToken(pingboardId string, pingboardSecret string) bool {
	// get auth token using client credentials
	response, err := c.restClient.R().
		SetQueryParams(map[string]string{"grant_type": "client_credentials"}).
		SetBody(fmt.Sprintf("{\"client_id\": \"%s\", \"client_secret\": \"%s\"}", pingboardId, pingboardSecret)).
		Post("https://app.pingboard.com/oauth/token")
	tokenResult := credentialsResponse{}
	if !c.pingboardResponse(response, err, "token", &tokenResult, func() bool {
		return tokenResult.Token != "" && tokenResult.SecondsRemaining != 0
	}) {
		c.pluginAPI.LogError("Failed to obtain auth token")
		return false
	}

	c.restClient.SetAuthToken(tokenResult.Token)
	return true
}


func (c *Client) resolveDepartment(user userResponse, departmentsById map[string]string) string {
	// We consider the user's department to be the first DepartmentId in the user's Links (if any) for which:
	// * groups/<departmentId> returns a single Group and
	// * the Id of that Group equals <departmentId>

	if user.Links.DepartmentIds == nil || len(user.Links.DepartmentIds) == 0 {
		return ""
	}

	departmentId := user.Links.DepartmentIds[0]
	department, found := departmentsById[departmentId]
	if found {
		return department
	}

	response, err := c.restClient.R().
		Get(fmt.Sprintf("https://app.pingboard.com/api/v2/groups/%s", departmentId))
	departmentResult := groupsResponse{}
	if !c.pingboardResponse(response, err, "department", &departmentResult, func() bool {
		return len(departmentResult.Groups) == 1 && departmentResult.Groups[0].Id == departmentId
	}) {
		return ""
	}
	c.pluginAPI.LogDebug(fmt.Sprintf("Found department with id %s, name %s",
		departmentId, departmentResult.Groups[0].Name))
	department = departmentResult.Groups[0].Name
	departmentsById[departmentId] = department

	return department
}

func (c *Client) FetchCompany() *Company {
	response, err := c.restClient.R().
		Get("https://app.pingboard.com/api/v2/companies/my_company")
	companiesResult := companiesResponse{}
	if !c.pingboardResponse(response, err, "companies", &companiesResult, func() bool {
		return len(companiesResult.Companies) == 1
	}) {
		return nil
	}
	company := companiesResult.Companies[0]
	c.pluginAPI.LogDebug(fmt.Sprintf("Pingboard query: Found company %s with sub-domain %s", company.Name, company.Domain))
	return &Company{
		Name:   company.Name,
		Domain: company.Domain,
	}
}

func (c *Client) FetchUsers() map[string]User {
	usersById := map[string]User{}
	departmentsById := map[string]string{}

	pageCount := 0
	for page := 1; pageCount == 0 || page <= pageCount; page += 1 {
		response, err := c.restClient.R().
			SetQueryParams(map[string]string{"page_size": "200", "page": fmt.Sprintf("%d", page)}).
			Get("https://app.pingboard.com/api/v2/users")
		usersResult := usersResponse{}
		if !c.pingboardResponse(response, err, "users", &usersResult, func() bool {
			return usersResult.Meta.Users.Page == page && len(usersResult.Users) > 0
		}) {
			return nil
		}
		c.pluginAPI.LogDebug(fmt.Sprintf("Pingboard query: got %d users (page %d)", len(usersResult.Users), page))
		pageCount = usersResult.Meta.Users.PageCount
		for _, user := range usersResult.Users {
			department := c.resolveDepartment(user, departmentsById)
			if department == "" {
				department = "(unknown department)"
			}
			c.pluginAPI.LogDebug(fmt.Sprintf("Found Pingboard user with "+
				"email %s, id %s, started %s, phone %s, title %s, manager id %d, department %s",
				user.Email, user.Id, user.StartDate, user.Phone, user.JobTitle, user.ReportsToId, department))
			reportsToId := ""
			if user.ReportsToId != 0 {
				reportsToId = strconv.Itoa(user.ReportsToId)
			}
			usersById[user.Id] = User{
				Id:          user.Id,
				StartDate:   user.StartDate,
				Email:       user.Email,
				Phone:       user.Phone,
				JobTitle:    user.JobTitle,
				ReportsToId: reportsToId,
				Department:  department,
			}
		}
	}
	c.pluginAPI.LogInfo(fmt.Sprintf("Found %d Pingboard users", len(usersById)))

	return usersById
}
