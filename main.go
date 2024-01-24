package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

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

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200"}, // Replace with your frontend's URL
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	handler := c.Handler(r)

	http.ListenAndServe(":8080", handler)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	response := JSONResponse{Success: true, Message: "Redirect to GitHub for authentication"}
	json.NewEncoder(w).Encode(response)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	code := r.URL.Query().Get("code")

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		jsonResponse(w, false, "Failed to exchange token: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	userResponse, err := client.Get("https://api.github.com/user")
	if err != nil {
		jsonResponse(w, false, "Failed to get user info: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}
	defer userResponse.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(userResponse.Body).Decode(&user); err != nil {
		jsonResponse(w, false, "Failed to decode user info: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, true, "User retrieved successfully", user, http.StatusOK)
}

func jsonResponse(w http.ResponseWriter, success bool, message string, data interface{}, statusCode int) {
	response := JSONResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
