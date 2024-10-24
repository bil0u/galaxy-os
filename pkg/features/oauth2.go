package features

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/bil0u/galaxy-os/pkg"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/json"
)

var (
	sessions         map[string]oauth2.Session = make(map[string]oauth2.Session)
	letters                                    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	loginTemplate                              = `<button><a href="%s">login</a></button>`
	loggedInTemplate                           = `"user:<br />%s<br />connections: <br />%s"`
)

func init() {
	pkg.RegisterFeature[Oauth2Feature]("oauth2")
}

type Oauth2Feature struct {
	Enabled    bool              `toml:"enabled"`
	BaseURL    string            `toml:"base_url"`
	ExposePort int               `toml:"port"`
	Routes     map[string]string `toml:"routes"`
}

func (f Oauth2Feature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "OAuth2",
		discord.LocaleFrench:    "OAuth2",
	}
}

func (f Oauth2Feature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "OAuth2 login flow",
		discord.LocaleFrench:    "Flux de connexion OAuth2",
	}
}

func (f Oauth2Feature) IsEnabled() bool {
	return f.Enabled
}

func (f Oauth2Feature) IsProperlyConfigured() error {
	var errs []error
	if f.ExposePort == 0 {
		errs = append(errs, fmt.Errorf("port is required"))
	}
	if f.Routes == nil || len(f.Routes) == 0 {
		errs = append(errs, fmt.Errorf("routes are required"))
	} else {
		// Check that "root", "authorize" and "redirect" routes are present
		if _, ok := f.Routes["root"]; !ok {
			errs = append(errs, fmt.Errorf("route 'root' is required"))
		}
		if _, ok := f.Routes["authorize"]; !ok {
			errs = append(errs, fmt.Errorf("route 'authorize' is required"))
		}
		if _, ok := f.Routes["redirect"]; !ok {
			errs = append(errs, fmt.Errorf("route 'redirect' is required"))
		}
	}
	if f.BaseURL == "" {
		errs = append(errs, fmt.Errorf("base_url is required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (f Oauth2Feature) Setup(bot *pkg.Bot) error {

	// Checking if feature is enabled
	if !f.Enabled {
		return nil
	}

	if f.ExposePort == 0 {
		f.ExposePort = 42000
	}

	// Setting default values
	if f.Routes == nil {
		f.Routes = map[string]string{
			"root":      "/oauth2",
			"authorize": "/oauth2/authorize",
			"redirect":  "/oauth2/redirect",
		}
	}

	// Checking if feature is properly configured
	if err := f.IsProperlyConfigured(); err != nil {
		return err
	}

	botConfig := bot.Config.Bot
	// Creating the OAuth2 client
	clientOpts := oauth2.WithRestClientConfigOpts(rest.WithHTTPClient(http.DefaultClient))
	bot.OAuthClient = oauth2.New(botConfig.ApplicationID, botConfig.ClientSecret, clientOpts)

	// Starting the web server
	mux := http.NewServeMux()
	mux.HandleFunc(f.Routes["root"], f.rootHandler(bot.OAuthClient))
	mux.HandleFunc(f.Routes["authorize"], f.authorizeHandler(bot.OAuthClient))
	mux.HandleFunc(f.Routes["redirect"], f.redirectHandler(bot.OAuthClient))
	go http.ListenAndServe(fmt.Sprintf(":%d", f.ExposePort), mux)
	return nil
}

type muxHandler func(http.ResponseWriter, *http.Request)

// rootHandler returns a function that handles the root route by checking if the user is logged in or not.
// If the user is logged in, it will display the user data and connections.
// If the user is not logged in, it will display a login button that redirects to the authorize route.
func (f Oauth2Feature) rootHandler(client oauth2.Client) muxHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		var body string

		// Retrieve the cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			slog.Info("No cookie found. User is not logged in")
			body = fmt.Sprintf(loginTemplate, f.Routes["authorize"])
		} else {
			slog.Info("Cookie found! User is logged in")

			// Retrieve the session from the sessions map
			session, ok := sessions[cookie.Value]
			if ok {
				// Session found, fetch user data
				var user *discord.OAuth2User
				user, err = client.GetUser(session)
				if err != nil {
					writeError(w, "error while getting user data", err)
					return
				}

				// Fetch connections data
				var connections []discord.Connection
				connections, err = client.GetConnections(session)
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
}

// authorizeHandler return a function that handles the OAuth2 authorization flow by redirecting the user to the Discord authorization URL.
// The user will be redirected back to the redirect URL after authorizing the application.
// It will ask for the following scopes:
// - User informations
// - Metadata
// - Third-party account connections
// - DM channels
// - Activities
func (f Oauth2Feature) authorizeHandler(client oauth2.Client) muxHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		params := oauth2.AuthorizationURLParams{
			RedirectURI: f.BaseURL + f.Routes["redirect"],
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
		slog.Info("Redirecting to the redirect URL")
		http.Redirect(w, r, client.GenerateAuthorizationURL(params), http.StatusSeeOther)
	}
}

// redirectHandler handles the OAuth2 redirect flow by starting a new session with the authorization code and state.
// The session is then stored in the sessions map and a cookie is set to keep track of the session.
func (f Oauth2Feature) redirectHandler(client oauth2.Client) muxHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			query = r.URL.Query()
			code  = query.Get("code")
			state = query.Get("state")
		)

		// If code and state are not empty, then it means the OAuth2 flow completed successfully
		// We can start a new session and store it in the sessions map
		if code != "" && state != "" {
			identifier := randStr(32)
			session, _, err := client.StartSession(code, state)
			if err != nil {
				writeError(w, "error while starting session", err)
				return
			}
			sessions[identifier] = session
			http.SetCookie(w, &http.Cookie{Name: "session_id", Value: identifier, Path: "/"})
			slog.Info("Session started: ", slog.Any("session", session))
		}

		slog.Info("Redirecting to root")
		http.Redirect(w, r, f.Routes["root"], http.StatusTemporaryRedirect)
	}
}

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

// ----------

func formatData(data any) []byte {
	var formatted []byte
	formatted, err := json.MarshalIndent(data, "<br />", "&ensp;")
	if err != nil {
		slog.Error("Failed to format data", slog.Any("data", data), slog.Any("error", err))
		return nil
	}
	return formatted
}
