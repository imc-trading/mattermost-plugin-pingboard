package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) writeApiError(w http.ResponseWriter, statusCode int, message string) {
	type Error struct {
		Error string `json:"error"`
	}
	w.WriteHeader(statusCode)
	b, err := json.Marshal(Error{Error: message})
	if err != nil {
		p.API.LogError("Failed to encode api error", "err", err)
		return
	}
	if _, err := w.Write(b); err != nil {
		p.API.LogError("Failed to write api error", "err", err)
		return
	}
}

func (p *Plugin) writeApiResponse(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		p.API.LogError("Failed to encode API result", "error", err)
		p.writeApiError(w, http.StatusInternalServerError, "failed to encode result")
		return
	}
	if _, err := w.Write(b); err != nil {
		p.API.LogError("Failed to write API result", "error", err)
		p.writeApiError(w, http.StatusInternalServerError, "failed to write result")
		return
	}
}

func (p *Plugin) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	usernames, ok := r.URL.Query()["username"]
	if !ok || len(usernames) != 1 {
		p.API.LogDebug("Returning bad request (malformed username param)", "usernames", usernames)
		p.writeApiError(w, http.StatusBadRequest, "specify one username")
		return
	}
	username := strings.ToLower(usernames[0])

	if p.usersByUsername == nil {
		p.API.LogDebug("Returning not found for " + username + " (no pingboard data)")
		http.NotFound(w, r)
		return
	}
	user, found := p.usersByUsername[username]
	if !found {
		p.API.LogDebug("Returning not found for " + username + " (unknown pingboard user)")
		http.NotFound(w, r)
		return
	}
	p.API.LogDebug("Returning user data for " + username)
	p.writeApiResponse(w, user)
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if userID := r.Header.Get("Mattermost-User-ID"); userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch path := r.URL.Path; path {
	case "/user":
		p.handleGetUser(w, r)
	default:
		http.NotFound(w, r)
	}
}
