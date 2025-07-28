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
	"/api/routes": handler.ApiHandler,
}

func main() {
	config.LoadConfig()
	db.InitDB()
	certs.RenewCerts()
	certs.CreateCertificates()

	router := http.NewServeMux()
	for key, value := range RedirectIgnores {
		router.HandleFunc(key, value)
	}
	router.HandleFunc("/", handler.ProxyHandler)

	loggedRouter := loggingMiddleware(router)

	serve(loggedRouter)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s from %s, to %s", r.Method, r.URL.Path, r.RemoteAddr, r.Host)
		next.ServeHTTP(w, r)
	})
}

func serve(router http.Handler) {
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpHandler(),
	}

	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: router,
		TLSConfig: &tls.Config{
			Certificates: certs.GetCertificates(),
		},
	}

	log.Fatal(httpServer.ListenAndServe())
	log.Fatal(httpsServer.ListenAndServeTLS("", ""))
}

func httpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Dont redirect api routes
		for path, handler := range RedirectIgnores {
			if path == r.URL.Path {
				handler(w, r)
				return
			}
		}

		// Get route from database
		host := r.Host
		route, err := db.GetRoute(host)
		if err != nil {
			http.Error(w, "Route not found", http.StatusNotFound)
			return
		}

		// If route should be https -> redirect
		if route.UseHTTPS {
			host := strings.Split(r.Host, ":")[0]
			target := "https://" + host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}

		handler.ProxyHandler(w, r)

	})

}
