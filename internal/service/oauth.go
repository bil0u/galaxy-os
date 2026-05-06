package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bil0u/galaxy-os/internal/platform"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

var (
	routeRoot      = "/oauth2"
	routeAuthorize = "/oauth2/authorize"
	routeRedirect  = "/oauth2/redirect"
)

// OAuthService wraps disgo's OAuth2 client as a platform.Service and platform.OAuthProvider.
type OAuthService struct {
	applicationID snowflake.ID
	clientSecret  string
	baseURL       string
	exposePort    int

	client   *oauth2.Client
	listener net.Listener
}

// NewOAuthService creates an OAuthService with the required configuration.
func NewOAuthService(applicationID snowflake.ID, clientSecret, baseURL string) *OAuthService {
	return &OAuthService{
		applicationID: applicationID,
		clientSecret:  clientSecret,
		baseURL:       baseURL,
		exposePort:    42000,
	}
}

func (s *OAuthService) Name() string { return "oauth" }

func (s *OAuthService) Start(_ context.Context) error {
	clientOpts := oauth2.WithRestClientConfigOpts(rest.WithHTTPClient(http.DefaultClient))
	s.client = oauth2.New(s.applicationID, s.clientSecret, clientOpts)

	mux := http.NewServeMux()
	mux.HandleFunc(routeRoot, s.rootHandler)
	mux.HandleFunc(routeAuthorize, s.authorizeHandler)
	mux.HandleFunc(routeRedirect, s.redirectHandler)

	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.exposePort))
	if err != nil {
		return fmt.Errorf("listening on port %d: %w", s.exposePort, err)
	}

	go func() {
		if serveErr := http.Serve(s.listener, mux); serveErr != nil && serveErr != http.ErrServerClosed {
			slog.Error("oauth http server error", slog.Any("error", serveErr))
		}
	}()
	return nil
}

func (s *OAuthService) Health(_ context.Context) platform.Health {
	status := platform.StatusDown
	if s.client != nil && s.listener != nil {
		status = platform.StatusUp
	}
	return platform.Health{
		Name:   s.Name(),
		Status: status,
	}
}

func (s *OAuthService) Stop(_ context.Context) error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Client returns the underlying OAuth2 client. Features type-assert
// OAuthProvider to *OAuthService to access this.
func (s *OAuthService) Client() *oauth2.Client {
	return s.client
}

// Handlers

var (
	sessionsMu       sync.RWMutex
	sessions         = make(map[string]oauth2.Session)
	loginTemplate    = `<button><a href="%s">login</a></button>`
	loggedInTemplate = `"user:<br />%s<br />connections: <br />%s"`
)

func (s *OAuthService) rootHandler(w http.ResponseWriter, r *http.Request) {
	var body string

	cookie, err := r.Cookie("session_id")
	if err != nil {
		body = fmt.Sprintf(loginTemplate, routeAuthorize)
	} else {
		sessionsMu.RLock()
		session, ok := sessions[cookie.Value]
		sessionsMu.RUnlock()
		if ok {
			var user *discord.OAuth2User
			user, err = s.client.GetUser(session)
			if err != nil {
				writeError(w, "error while getting user data", err)
				return
			}

			var connections []discord.Connection
			connections, err = s.client.GetConnections(session)
			if err != nil {
				writeError(w, "error while getting connections data", err)
				return
			}

			userJSON := formatData(user)
			connectionsJSON := formatData(connections)

			body = fmt.Sprintf(loggedInTemplate, userJSON, connectionsJSON)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (s *OAuthService) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	params := oauth2.AuthorizationURLParams{
		RedirectURI: s.baseURL + routeRedirect,
		Scopes: []discord.OAuth2Scope{
			discord.OAuth2ScopeIdentify,
			discord.OAuth2ScopeRoleConnectionsWrite,
			discord.OAuth2ScopeConnections,
			discord.OAuth2ScopeGDMJoin,
		},
	}
	http.Redirect(w, r, s.client.GenerateAuthorizationURL(params), http.StatusSeeOther)
}

func (s *OAuthService) redirectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query = r.URL.Query()
		code  = query.Get("code")
		state = query.Get("state")
	)

	if code != "" && state != "" {
		identifier := randStr(32)
		session, _, err := s.client.StartSession(code, state)
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

// Utils

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func writeError(w http.ResponseWriter, text string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(text + ": " + err.Error()))
}

func randStr(n int) string {
	var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}

func formatData(data any) []byte {
	var formatted []byte
	formatted, err := json.MarshalIndent(data, "<br />", "&ensp;")
	if err != nil {
		slog.Error("Failed to format data", slog.Any("data", data), slog.Any("error", err))
		return nil
	}
	return formatted
}
