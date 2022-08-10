package main

import (
	"flag"
	"fmt"
	gologin "github.com/dghubble/gologin"
	gologin2 "github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/dghubble/sessions"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
	"log"
	"net/http"
	"os"
)

const (
	sessionName     = "example-github-app"
	sessionSecret   = "example cookie signing secret"
	sessionUserKey  = "githubID"
	sessionUsername = "githubUsername"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// Config configures the main ServeMux.
type Config struct {
	GithubClientID     string
	GithubClientSecret string
}

// New returns a new ServeMux with app routes.
func New(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", profileHandler)
	mux.HandleFunc("/logout", logoutHandler)

	oauth2Config := &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		RedirectURL:  "http://localhost:3000/github/callback",
		Endpoint:     githubOAuth2.Endpoint,
	}

	stateConfig := gologin.DebugOnlyCookieConfig
	mux.Handle("/github/login", github.StateHandler(gologin2.CookieConfig(stateConfig), github.LoginHandler(oauth2Config, nil)))
	mux.Handle("/github/callback", github.StateHandler(gologin2.CookieConfig(stateConfig), github.CallbackHandler(oauth2Config, issueSession(), nil)))
	return mux
}

func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		githubUser, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session := sessionStore.New(sessionName)
		session.Values[sessionUserKey] = *githubUser.ID
		session.Values[sessionUsername] = *githubUser.Login
		error1 := session.Save(w)
		if error1 != nil {
			return
		}
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func profileHandler(w http.ResponseWriter, req *http.Request) {
	session, err := sessionStore.Get(req, sessionName)
	if err != nil {
		// welcome with login button
		page, _ := os.ReadFile("index.html")
		_, error1 := fmt.Fprintf(w, string(page))
		if error1 != nil {
			return
		}
		return
	}

	_, error2 := fmt.Fprintf(w, `<p>You are logged in %s!</p><form action="/logout" method="post">
							<input type="submit" value="Logout"></form>`, session.Values[sessionUsername])
	if error2 != nil {
		return
	}
}

func logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		sessionStore.Destroy(w, sessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func main() {
	const address = "localhost:3000"
	config := &Config{
		GithubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
	clientID := flag.String("client-id", "", "Github Client ID")
	clientSecret := flag.String("client-secret", "", "Github Client Secret")
	flag.Parse()
	if *clientID != "" {
		config.GithubClientID = *clientID
	}
	if *clientSecret != "" {
		config.GithubClientSecret = *clientSecret
	}
	if config.GithubClientID == "" {
		log.Fatal("Missing Github Client ID")
	}
	if config.GithubClientSecret == "" {
		log.Fatal("Missing Github Client Secret")
	}

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
