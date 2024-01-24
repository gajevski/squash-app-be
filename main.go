package main

import (
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
