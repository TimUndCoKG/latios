package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"strings"

	"github.com/timsalokat/latios_proxy/db"
	"github.com/timsalokat/latios_proxy/handler"
)

var RedirectIgnores = map[string]func(http.ResponseWriter, *http.Request){
	"/api/routes": handler.ApiHandler,
}

func main() {
	db.InitDB()
	certs.InitRenewal()
	router := http.NewServeMux()
	for key, value := range RedirectIgnores {
		router.HandleFunc(key, value)
	}
	router.HandleFunc("/", handler.ProxyHandler)

	serve(router)
}

func serve(router http.Handler) {
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpHandler(router),
	}

	httpsServer := &http.Server{
		Addr:      ":443",
		Handler:   router,
		TLSConfig: &tls.Config{
			//TODO TBD
		},
	}

	log.Fatal(httpServer.ListenAndServe())
	log.Fatal(httpsServer.ListenAndServeTLS("", ""))
}

func httpHandler(baseRouter http.Handler) http.Handler {
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
