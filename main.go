package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type User struct {
	ID       string
	Username string
}

var oauth2Config = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("CALLBACK_REDIRECT"),
	Scopes:       []string{"read:user"},
	Endpoint:     github.Endpoint,
}
var jwtKey = []byte(os.Getenv("JWT_KEY"))

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := mux.NewRouter()
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/callback", callbackHandler)
	http.ListenAndServe(":8080", r)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200"},
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

	var user User
	if err := json.NewDecoder(userResponse.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jwtToken, err := generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": jwtToken})
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

func generateToken(user User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &jwt.StandardClaims{
		Subject:   user.ID,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	return tokenString, err
}
