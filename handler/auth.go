package handler

import (
	_ "embed"
	"fmt"
	"html/template"
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
		if r.URL.Path == "/latios-api/login" || r.URL.Path == "/latios-api/health" {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("[AUTH] Checking authentication for host: %s path: %s", r.Host, r.URL.Path)

		if !routeRequiresAuth(r.Host) {
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
				loginURL := fmt.Sprintf("/latios-api/login?redirect=%s", url.QueryEscape(r.URL.String()))
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

	// Show simple login form
	case http.MethodGet:
		redirect := r.URL.Query().Get("redirect")
		loginPage(w, r, redirect, "")

	// Try to login user via username and password
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		redirect := r.Form.Get("redirect")

		isSecure := r.TLS != nil

		if validateCredentials(username, password) {
			// Set cookie
			token, err := generateToken(username)
			if err != nil {
				log.Printf("[AUHT] Couldnt create token for user: %s", username)
				loginPage(w, r, redirect, "Couldnt create token")
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
			log.Printf("[AUTH] Invalid credentials for user %s", username)
			loginPage(w, r, redirect, "Invalid username or password")
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

//go:embed templates/login.html
var loginHTML string
var login_template = template.Must(template.New("login.html").Parse(loginHTML))

type LoginData struct {
	Redirect string
	ErrorMsg string
}

func loginPage(w http.ResponseWriter, r *http.Request, redirect string, err string) {
	data := LoginData{
		Redirect: redirect,
		ErrorMsg: err,
	}

	if err := login_template.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
