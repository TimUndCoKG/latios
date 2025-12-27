package handler

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/timsalokat/latios_proxy/db"
	"golang.org/x/crypto/bcrypt"
)

var authCookieName = "latios_auth"
var jwtKey = []byte(os.Getenv("LATIOS_SECRET_KEY"))

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Redirect string `json:"redirect"`
}

// Simple check if route requires security
func routeRequiresAuth(host string) bool {
	route, err := db.GetRoute(host)
	if err != nil {
		return true
	}
	return route.EnforceAuth
}

func validateCredentials(username, password string) bool {
	var user db.User
	if err := db.Client.Where("username = ?", username).First(&user).Error; err != nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func generateToken(username string) (string, error) {
	exporationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exporationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Middleware to check authentication and redirect to login if needed
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If the path is the login page or the healthcheck, skip auth
		if strings.HasPrefix(r.URL.Path, "/latios/assets/") ||
			r.URL.Path == "/latios/login" ||
			r.URL.Path == "/latios-api/login" ||
			r.URL.Path == "/latios-api/health" {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("[AUTH] Checking authentication for host: %s path: %s", r.Host, r.URL.Path)

		if !strings.HasPrefix(r.URL.Path, "/latios") && !routeRequiresAuth(r.Host) {
			log.Printf("[AUTH] Route does not require auth, proceeding")
			next.ServeHTTP(w, r)
			return
		}

		tokenString := ""
		if cookie, err := r.Cookie(authCookieName); err == nil {
			tokenString = cookie.Value
		} else {
			header := r.Header.Get("Authorization")
			if strings.HasPrefix(header, "Bearer ") {
				tokenString = strings.TrimPrefix(header, "Bearer ")
			}
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/latios-api/") {
				loginURL := fmt.Sprintf("/latios/login?redirect=%s", url.QueryEscape(r.URL.String()))
				http.Redirect(w, r, loginURL, http.StatusFound)
			} else {
				http.Error(w, "Unatuhorized", http.StatusUnauthorized)
			}
			return
		}

		log.Printf("[AUTH] Authenticated user, proceeding")
		next.ServeHTTP(w, r)

	})
}

// Handle the login page GET and POST
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// Try to login user via username and password
	case http.MethodPost:
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[AUTH] Error decoding login json: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		username := req.Username
		password := req.Password
		redirect := req.Redirect

		isSecure := r.TLS != nil

		if validateCredentials(username, password) {
			// Set cookie
			token, err := generateToken(username)
			if err != nil {
				log.Printf("[AUTH] Couldnt create token for user: %s", username)
				gotoLogin(w, r)
			}

			http.SetCookie(w, &http.Cookie{
				Name:     authCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   isSecure,
				Expires:  time.Now().Add(24 * time.Hour),
				SameSite: http.SameSiteLaxMode, // prevents csrf attacks (not sure if this option doesnt deny my sso)
			})

			log.Printf("[AUTH] User %s logged in successfully, redirecting to %s", username, redirect)
			if redirect == "" {
				redirect = "/"
			}
			http.Redirect(w, r, redirect, http.StatusFound)

		} else {
			log.Printf("[AUTH] Invalid credentials for user: %s", username)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func gotoLogin(w http.ResponseWriter, r *http.Request) {
	loginURL := fmt.Sprintf("/latios/login?redirect=%s", url.QueryEscape(r.URL.String()))
	http.Redirect(w, r, loginURL, http.StatusFound)
}
