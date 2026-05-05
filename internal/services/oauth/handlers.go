package oauth

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/oauth2"
)

var (
	sessionsMu       sync.RWMutex
	sessions         = make(map[string]oauth2.Session)
	loginTemplate    = `<button><a href="%s">login</a></button>`
	loggedInTemplate = `"user:<br />%s<br />connections: <br />%s"`
)

// rootHandler returns a function that handles the root route by checking if the user is logged in or not.
// If the user is logged in, it will display the user data and connections.
// If the user is not logged in, it will display a login button that redirects to the authorize route.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	var body string

	// Retrieve the cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		body = fmt.Sprintf(loginTemplate, routeAuthorize)
	} else {
		sessionsMu.RLock()
		session, ok := sessions[cookie.Value]
		sessionsMu.RUnlock()
		if ok {
			// Session found, fetch user data
			var user *discord.OAuth2User
			user, err = (*Client).GetUser(session)
			if err != nil {
				writeError(w, "error while getting user data", err)
				return
			}

			// Fetch connections data
			var connections []discord.Connection
			connections, err = (*Client).GetConnections(session)
			if err != nil {
				writeError(w, "error while getting connections data", err)
				return
			}

			// Format the data
			userJSON := formatData(user)
			connectionsJSON := formatData(connections)

			body = fmt.Sprintf(loggedInTemplate, userJSON, connectionsJSON)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// authorizeHandler return a function that handles the OAuth2 authorization flow by redirecting the user to the Discord authorization URL.
// The user will be redirected back to the redirect URL after authorizing the application.
// It will ask for the following scopes:
// - User informations
// - Metadata
// - Third-party account connections
// - DM channels
// - Activities
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	params := oauth2.AuthorizationURLParams{
		RedirectURI: serverBaseURL + routeRedirect,
		Scopes: []discord.OAuth2Scope{
			// User informations
			discord.OAuth2ScopeIdentify,
			// Metadata
			discord.OAuth2ScopeRoleConnectionsWrite,
			// Third-party account connections
			discord.OAuth2ScopeConnections,
			// DM channels
			discord.OAuth2ScopeGDMJoin,
		},
	}
	http.Redirect(w, r, (*Client).GenerateAuthorizationURL(params), http.StatusSeeOther)
}

// redirectHandler handles the OAuth2 redirect flow by starting a new session with the authorization code and state.
// The session is then stored in the sessions map and a cookie is set to keep track of the session.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query = r.URL.Query()
		code  = query.Get("code")
		state = query.Get("state")
	)

	// If code and state are not empty, then it means the OAuth2 flow completed successfully
	// We can start a new session and store it in the sessions map
	if code != "" && state != "" {
		identifier := randStr(32)
		session, _, err := (*Client).StartSession(code, state)
		if err != nil {
			writeError(w, "error while starting session", err)
			return
		}
		sessionsMu.Lock()
		sessions[identifier] = session
		sessionsMu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: identifier, Path: "/"})
	}
	http.Redirect(w, r, routeRoot, http.StatusTemporaryRedirect)

}
