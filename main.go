package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var oauth2Config = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	RedirectURL:  "http://localhost:8080/callback",
	Scopes:       []string{"read:user"},
	Endpoint:     github.Endpoint,
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/callback", callbackHandler)
	http.ListenAndServe(":8080", r)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	userResponse, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userResponse.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(userResponse.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User Info: %v", user)
}
