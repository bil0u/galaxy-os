package oauth

import (
	"fmt"
	"net/http"

	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

var (
	oauthClient    oauth2.Client
	serverBaseURL  string
	exposePort     = 42000
	routeRoot      = "/oauth2"
	routeAuthorize = "/oauth2/authorize"
	routeRedirect  = "/oauth2/redirect"
)

func Init(applicationID snowflake.ID, clientSecret, baseURL string) oauth2.Client {
	serverBaseURL = baseURL
	clientOpts := oauth2.WithRestClientConfigOpts(rest.WithHTTPClient(http.DefaultClient))
	oauthClient = oauth2.New(applicationID, clientSecret, clientOpts)
	return oauthClient
}

func Client() oauth2.Client {
	return oauthClient
}

func Start() {
	mux := http.NewServeMux()
	mux.HandleFunc(routeRoot, rootHandler)
	mux.HandleFunc(routeAuthorize, authorizeHandler)
	mux.HandleFunc(routeRedirect, redirectHandler)

	go http.ListenAndServe(fmt.Sprintf(":%d", exposePort), mux)
}
