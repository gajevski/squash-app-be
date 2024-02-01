package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Racket   Racket `json:"racket"`
}

type Racket struct {
	Name                string `json:"name"`
	Image               string `json:"image"`
	PurchaseDate        string `json:"purchaseDate"`
	PlayedMatchesAmount int    `json:"playedMatchesAmount"`
	Grip                string `json:"grip"`
	String              string `json:"string"`
}

var (
	oauth2Config *oauth2.Config
	jwtKey       []byte
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("CALLBACK_REDIRECT"),
		Scopes:       []string{"read:user"},
		Endpoint:     github.Endpoint,
	}
	jwtKey = []byte(os.Getenv("JWT_KEY"))

	r := mux.NewRouter()
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/callback", callbackHandler)
	r.HandleFunc("/api/user", userHandler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	log.Fatal(http.ListenAndServe(":8080", handler))
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
		log.Printf("Error exchanging token: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	userResponse, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer userResponse.Body.Close()

	var user User
	if err := json.NewDecoder(userResponse.Body).Decode(&user); err != nil {
		log.Printf("Error decoding user info: %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	jwtToken, err := generateToken(user)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": jwtToken})
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := User{
		Username: "Mikolaj",
		ID:       1,
		Image:    "https://avatars.githubusercontent.com/u/29663156?v=4",
		Racket: Racket{
			Name:                "Wilson Hyper Hammer 120",
			Image:               "https://www.squashtime.pl/images/thumbs/640_720/WRT967700_wilson_01.jpg",
			PurchaseDate:        "October 2023",
			PlayedMatchesAmount: 26,
			Grip:                "Toalson Ultra Grip 3Pack Black",
			String:              "Default",
		},
	}
	json.NewEncoder(w).Encode(user)
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
		Subject:   strconv.FormatInt(user.ID, 10),
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	return tokenString, err
}
