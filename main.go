package main

import (
	"crypto/tls"
	"embed"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/timsalokat/latios_proxy/certs"
	"github.com/timsalokat/latios_proxy/config"
	"github.com/timsalokat/latios_proxy/db"
	"github.com/timsalokat/latios_proxy/handler"
)

//go:embed all:latios-frontend/dist
var content embed.FS

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

	// Register /latios-api and /latios
	handler.RegisterApiHandlers(router)
	if err := handler.RegisterFrontendHandlers(router, content); err != nil {
		log.Fatalf("[BOOT] Failed to register frontend handlers: %v", err)
	}

	// Register proxy handler
	log.Println("[ROUTER] Setting up default proxy handler for /")
	router.HandleFunc("/", handler.ProxyHandler)

	log.Println("[MIDDLEWARE] Adding request logging middleware...")
	loggedRouter := loggingMiddleware(router)
	secureRouter := handler.AuthMiddleware(loggedRouter)

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
		log.Printf("[HTTP] Request: %s %s Host=%s", r.Method, r.URL.Path, r.Host)

		route, err := db.GetRoute(r.Host)
		if err != nil {
			log.Printf("[DB] No HTTPS route for host=%s: %v", r.Host, err)
			router.ServeHTTP(w, r) // Pass to main router
			return
		}

		if route.UseHTTPS {
			target := "https://" + strings.Split(r.Host, ":")[0] + r.URL.RequestURI()
			log.Printf("[HTTP-REDIRECT] Redirecting to HTTPS: %s", target)
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}

		router.ServeHTTP(w, r)
	})
}
