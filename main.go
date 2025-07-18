package main

import (
	"log"
	"net/http"

	"github.com/timsalokat/latios_proxy/db"
	"github.com/timsalokat/latios_proxy/handlers"
	"github.com/timsalokat/latios_proxy/proxy"
	"github.com/timsalokat/latios_proxy/tls"
)

func main() {
	db.InitDB("routes.db")

	router := http.NewServeMux()

	// API endpoints
	handlers.RegisterRoutes(router)

	// Catch-all proxy handler
	router.HandleFunc("/", proxy.ProxyHandler)

	log.Fatal(http.ListenAndServe(":80", router))

	// TLS-enabled server with autocert
	tls.ServeWithTLS(router)
}
