package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/timsalokat/latios_proxy/certs"
	"github.com/timsalokat/latios_proxy/config"
	"github.com/timsalokat/latios_proxy/db"
	"github.com/timsalokat/latios_proxy/handler"
)

var RedirectIgnores = map[string]func(http.ResponseWriter, *http.Request){
	"/latios/routes": handler.RoutesApiHandler,
	"/latios/login":  handler.LoginHandler,
	"/latios/health": handler.HealthCheckHandler,
}

func main() {
	log.Println("[BOOT] Starting Latios proxy...")

	log.Println("[CONFIG] Loading configuration...")
	config.LoadConfig()

	log.Println("[DB] Initializing database...")
	db.InitDB()

	if os.Getenv("ENVIRONMENT") == "live" {
		log.Println("[CERTS] Renewing existing certificates...")
		certs.RenewCerts()

		log.Println("[CERTS] Creating new certificates (if needed)...")
		certs.CreateCertificates()
	}

	router := http.NewServeMux()
	log.Println("[MIDDLEWARE] Adding request logging middleware...")
	loggedRouter := loggingMiddleware(router)
	secureRouter := handler.AuthMiddleware(loggedRouter)

	log.Println("[ROUTER] Setting up routes...")
	for key, value := range RedirectIgnores {
		log.Printf("[ROUTER] Adding ignored redirect path: %s\n", key)
		router.HandleFunc(key, value)
	}
	router.HandleFunc("/", handler.ProxyHandler)

	log.Println("[SERVE] Starting HTTP and HTTPS servers...")
	serve(secureRouter)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[REQUEST] Method=%s Path=%s RemoteAddr=%s Host=%s Headers=%v",
			r.Method, r.URL.Path, r.RemoteAddr, r.Host, r.Header)
		next.ServeHTTP(w, r)
	})
}

func serve(router http.Handler) {
	if os.Getenv("ENVIRONMENT") == "live" {
		log.Println("[HTTPS] Preparing HTTPS server on :443")
		httpsServer := &http.Server{
			Addr:    ":443",
			Handler: router,
			TLSConfig: &tls.Config{
				Certificates: certs.GetCertificates(),
			},
		}

		go func() {
			// Start HTTPS server
			log.Println("[HTTPS] Starting HTTPS server on :443")
			if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
				log.Printf("[HTTPS] ERROR: %v\n", err)
			}
		}()
	}

	log.Println("[HTTP] Preparing HTTP server on :80")
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpHandler(router),
	}

	// Start HTTP redirect server
	log.Println("[HTTP] Starting HTTP server on :80 (redirect handler enabled)")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Printf("[HTTP] ERROR: %v\n", err)
	}

}

func httpHandler(router http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP-REDIRECT] Request: %s %s Host=%s", r.Method, r.URL.Path, r.Host)

		// Get route from database
		route, err := db.GetRoute(r.Host)
		if err != nil {
			log.Printf("[DB] No HTTPS route for host=%s: %v", r.Host, err)
			router.ServeHTTP(w, r) // just let router handle (includes /latios/*)
			return
		}

		if route.UseHTTPS {
			target := "https://" + strings.Split(r.Host, ":")[0] + r.URL.RequestURI()
			log.Printf("[HTTP-REDIRECT] Redirecting to HTTPS: %s", target)
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}

		// Otherwise, handle normally
		router.ServeHTTP(w, r)
	})
}
