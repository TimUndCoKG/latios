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

	"github.com/timsalokat/latios_proxy/db"
)

var authCookieName = "latios_auth"
var authHeaderName = "latios_auth"

// Simple check if route requires security
func routeRequiresAuth(host string) bool {
	route, err := db.GetRoute(host)
	if err != nil {
		return true
	}
	return route.EnforceAuth
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
			// Route doesn't need auth
			next.ServeHTTP(w, r)
			return
		}
		// Check if user is authenticated via cookie
		log.Printf("[AUTH] Checking authentication via cookie")

		cookie, err := r.Cookie(authCookieName)

		// TODO this is not an encrypted value. Make this proper api key
		if err != nil || cookie.Value != os.Getenv("LATIOS_AUTH_COOKIE") {

			log.Printf("[AUTH] Checking authentication via header")
			header := r.Header.Get(authHeaderName)

			// TODO this is bullshit, just because a request is of method GET does not mean it is made via a browser that has to be redirected
			if header == "" {
				if r.Method == http.MethodGet {
					log.Printf("[AUTH] Not authenticated, redirecting to login")
					loginURL := fmt.Sprintf("/latios-api/login?redirect=%s", url.QueryEscape(r.URL.String()))
					http.Redirect(w, r, loginURL, http.StatusFound)
					return
				} else {
					log.Println("[AUTH] Invalid Authorization header")
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			// TODO simplify these two
			if !strings.HasPrefix(header, "Bearer ") {
				log.Printf("[AUTH] Invalid Authorization format")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token != os.Getenv("LATIOS_AUTH_TOKEN") {
				log.Println("[AUTH] Invalid Authorization header")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Authenticated, proceed
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
			http.SetCookie(w, &http.Cookie{
				Name:     authCookieName,
				Value:    os.Getenv("LATIOS_AUTH_COOKIE"),
				Path:     "/",
				HttpOnly: true,
				Secure:   isSecure,
				Expires:  time.Now().Add(4 * time.Hour),
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

func validateCredentials(username, password string) bool {
	expectedUser := os.Getenv("FALLBACK_USER")
	expectedPass := os.Getenv("FALLBACK_PASSWORD")
	return username == expectedUser && password == expectedPass
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
