package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/imc/mattermost-plugin-pingboard/server/pingboard"
)

type pingboardData struct {
	company   *pingboard.Company
	usersById map[string]pingboard.User
}

var dateExpr = regexp.MustCompile(`([0-9]{4})-([0-9]{2})-([0-9]{2})`)

var emailRetainedChars = regexp.MustCompile(`[^a-z0-9@.]`)

// Remove special characters from emails and flip to lowercase; this helps to ensure
// that mattermost and pingboard agree on the email string. So the following
// email addresses are considered a match:
// Some-Body.with+Chars@Domain+123.com
// somebody.withchars@domain123.com
func normalisedEmail(email string) string {
	return emailRetainedChars.ReplaceAllString(strings.ToLower(email), "")
}

func (p *Plugin) getMattermostUsernamesByNormalisedEmail() map[string]string {
	mmUsernamesByNormalisedEmail := map[string]string{}

	page := 0
	for {
		mmUsers, err := p.API.GetUsers(&model.UserGetOptions{
			Page:    page,
			PerPage: 500,
		})
		if err != nil {
			p.API.LogError("Failed to get mattermost users", err)
		}
		if len(mmUsers) == 0 {
			break
		}
		p.API.LogDebug(fmt.Sprintf("Scan mattermost users: got %d users (page %d)", len(mmUsers), page))
		page += 1
		for _, mattermostUser := range mmUsers {
			mmEmail := normalisedEmail(mattermostUser.Email)
			if _, exists := mmUsernamesByNormalisedEmail[mmEmail]; exists {
				p.API.LogError(fmt.Sprintf("Found multiple mattermost users with (normalised) email %s", mmEmail))
				return nil
			}
			p.API.LogDebug(fmt.Sprintf("Found mattermost user %s with normalised email %s",
				mattermostUser.Username, mmEmail))
			mmUsernamesByNormalisedEmail[mmEmail] = mattermostUser.Username
		}
	}
	p.API.LogInfo(fmt.Sprintf("Found %d mattermost users", len(mmUsernamesByNormalisedEmail)))

	return mmUsernamesByNormalisedEmail
}

func (p *Plugin) fetchPingboardData(apiID string, apiSecret string) *pingboardData {
	pbClient := pingboard.NewClient(p.API, apiID, apiSecret)

	if pbClient == nil {
		return nil
	}

	company := pbClient.FetchCompany()
	if company == nil {
		return nil
	}

	pbUsersById := pbClient.FetchUsers()
	if pbUsersById == nil {
		return nil
	}

	return &pingboardData{
		company:   company,
		usersById: pbUsersById,
	}
}

func (p *Plugin) resolveUsers(pbData *pingboardData, mmUsernamesByNormalisedEmail map[string]string) map[string]User {
	usersByUsername := map[string]User{}
	pbNormalisedEmails := map[string]bool{}
	for _, pbUser := range pbData.usersById {
		pbUserNormalisedEmail := normalisedEmail(pbUser.Email)

		mmUsername, found := mmUsernamesByNormalisedEmail[pbUserNormalisedEmail]
		if !found {
			p.API.LogDebug(fmt.Sprintf("Ignoring Pingboard user with normalised email %s (no matching mattermost user)",
				pbUserNormalisedEmail))
			continue
		}

		if _, exists := pbNormalisedEmails[pbUserNormalisedEmail]; exists {
			p.API.LogError(fmt.Sprintf("Found multiple Pingboard users with (normalised) email %s",
				pbUserNormalisedEmail))
			return nil
		}
		pbNormalisedEmails[pbUserNormalisedEmail] = true

		p.API.LogDebug(fmt.Sprintf("Recording data for user %s matched by normalised email %s",
			mmUsername, pbUserNormalisedEmail))

		startYear := 0
		startMonth := 0
		startDay := 0
		dateParts := dateExpr.FindStringSubmatch(pbUser.StartDate)
		if dateParts != nil {
			startYear, _ = strconv.Atoi(dateParts[1])
			startMonth, _ = strconv.Atoi(dateParts[2])
			startDay, _ = strconv.Atoi(dateParts[3])
		}

		manager := ""
		managerId := pbUser.ReportsToId
		if managerId != "" {
			if managerUser, found := pbData.usersById[managerId]; found {
				managerEmail := normalisedEmail(managerUser.Email)
				if manager, found = mmUsernamesByNormalisedEmail[managerEmail]; found {
					p.API.LogDebug(fmt.Sprintf("User %s matched to manager %s by normalised email %s",
						mmUsername, manager, managerEmail))
				} else {
					p.API.LogDebug(fmt.Sprintf("User %s has manager with unmatched normalised email %s",
						mmUsername, managerEmail))
				}
			} else {
				p.API.LogDebug(fmt.Sprintf("User %s has manager with unknown Pingboard ID %s",
					mmUsername, managerId))
			}
		}

		newUser := User{
			Id:         pbUser.Id,
			Email:      pbUser.Email,
			Url:        fmt.Sprintf("https://%s.pingboard.com/users/%s", pbData.company.Domain, pbUser.Id),
			StartYear:  startYear,
			StartMonth: startMonth,
			StartDay:   startDay,
			Phone:      pbUser.Phone,
			JobTitle:   pbUser.JobTitle,
			Department: pbUser.Department,
			Manager:    manager,
		}

		usersByUsername[mmUsername] = newUser
	}

	return usersByUsername
}

func (p *Plugin) refreshData() {
	p.refreshLock.Lock()
	defer p.refreshLock.Unlock()

	if p.refreshTimer == nil {
		return
	}
	p.refreshTimer.Stop()

	config := p.getConfiguration()
	clientId := config.PingboardApiId
	clientSecret := os.Getenv("MM_PLUGIN_PINGBOARD_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = config.PingboardApiSecret
	}

	if clientId == "" || clientSecret == "" {
		p.API.LogInfo("No Pingboard client configuration")
		// do not schedule more attempts (config change will already trigger a refresh)
		return
	}

	// always schedule a later attempt even if we fail with errors below
	p.refreshTimer = time.AfterFunc(time.Duration(6)*time.Hour, p.refreshData)

	p.API.LogInfo("Refreshing data...")

	// Index all mattermost users by normalised email address
	mmUsernamesByNormalisedEmail := p.getMattermostUsernamesByNormalisedEmail()
	if mmUsernamesByNormalisedEmail == nil {
		return
	}

	// Get data from pingboard
	pbData := p.fetchPingboardData(clientId, clientSecret)
	if pbData == nil {
		return
	}

	// Assemble final info by usernames
	usersByUsername := p.resolveUsers(pbData, mmUsernamesByNormalisedEmail)
	if usersByUsername == nil {
		return
	}

	p.usersByUsername = usersByUsername
}
