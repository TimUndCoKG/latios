package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"strings"

	"github.com/timsalokat/latios_proxy/certs"
	"github.com/timsalokat/latios_proxy/config"
	"github.com/timsalokat/latios_proxy/db"
	"github.com/timsalokat/latios_proxy/handler"
)

var RedirectIgnores = map[string]func(http.ResponseWriter, *http.Request){
	"/latios/routes": handler.RoutesApiHandler,
	"/login":         handler.LoginHandler,
}

func main() {
	log.Println("[BOOT] Starting Latios proxy...")

	log.Println("[CONFIG] Loading configuration...")
	config.LoadConfig()

	log.Println("[DB] Initializing database...")
	db.InitDB()

	log.Println("[CERTS] Renewing existing certificates...")
	certs.RenewCerts()

	log.Println("[CERTS] Creating new certificates (if needed)...")
	certs.CreateCertificates()

	log.Println("[ROUTER] Setting up routes...")
	router := http.NewServeMux()
	for key, value := range RedirectIgnores {
		log.Printf("[ROUTER] Adding ignored redirect path: %s\n", key)
		router.HandleFunc(key, value)
	}
	router.HandleFunc("/", handler.ProxyHandler)

	log.Println("[MIDDLEWARE] Adding request logging middleware...")
	loggedRouter := loggingMiddleware(router)
	secureRouer := handler.AuthMiddleware(loggedRouter)

	log.Println("[SERVE] Starting HTTP and HTTPS servers...")
	serve(secureRouer)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[REQUEST] Method=%s Path=%s RemoteAddr=%s Host=%s Headers=%v",
			r.Method, r.URL.Path, r.RemoteAddr, r.Host, r.Header)
		next.ServeHTTP(w, r)
	})
}

func serve(router http.Handler) {
	log.Println("[HTTP] Preparing HTTP server on :80")
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpHandler(),
	}

	log.Println("[HTTPS] Preparing HTTPS server on :443")
	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: router,
		TLSConfig: &tls.Config{
			Certificates: certs.GetCertificates(),
		},
	}

	// Start HTTP redirect server
	go func() {
		log.Println("[HTTP] Starting HTTP server on :80 (redirect handler enabled)")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Printf("[HTTP] ERROR: %v\n", err)
		}
	}()

	// Start HTTPS server
	log.Println("[HTTPS] Starting HTTPS server on :443")
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		log.Printf("[HTTPS] ERROR: %v\n", err)
	}
}

func httpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP-HANDLER] Incoming request for %s %s Host=%s", r.Method, r.URL.Path, r.Host)

		// Don't redirect API routes
		for path, handler := range RedirectIgnores {
			if path == r.URL.Path {
				log.Printf("[HTTP-HANDLER] Path %s is in RedirectIgnores. Handling directly.", path)
				handler(w, r)
				return
			}
		}

		// Get route from database
		host := r.Host
		log.Printf("[HTTP-HANDLER] Looking up route for host: %s", host)
		route, err := db.GetRoute(host)
		if err != nil {
			log.Printf("[DB] ERROR: Route not found for host=%s: %v", host, err)
			http.Error(w, "Route not found", http.StatusNotFound)
			return
		}
		log.Printf("[HTTP-HANDLER] Found route: %+v", route)

		// If route should use HTTPS -> redirect
		if route.UseHTTPS {
			host := strings.Split(r.Host, ":")[0]
			target := "https://" + host + r.URL.RequestURI()
			log.Printf("[HTTP-HANDLER] Redirecting to HTTPS: %s", target)
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}

		log.Printf("[HTTP-HANDLER] Forwarding to ProxyHandler (no HTTPS redirect)")
		handler.ProxyHandler(w, r)
	})
}
