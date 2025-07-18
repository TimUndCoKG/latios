// tls/tls.go
package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/timsalokat/latios_proxy/db"
	"golang.org/x/crypto/acme/autocert"
)

func ServeWithTLS(handler http.Handler) {
	m := &autocert.Manager{
		Cache:      autocert.DirCache(".certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: dynamicHostPolicy,
	}

	// Custom HTTP handler to skip redirect for /api/routes
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/routes" {
			handler.ServeHTTP(w, r)
			return
		}
		m.HTTPHandler(nil).ServeHTTP(w, r)
	})

	httpsSrv := &http.Server{
		Addr:      ":443",
		Handler:   handler,
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}

	// Redirect HTTP to HTTPS, except for /api/routes
	go func() {
		log.Fatal(http.ListenAndServe(":80", httpHandler))
	}()

	log.Println("Listening on :443 with TLS")
	log.Fatal(httpsSrv.ListenAndServeTLS("", ""))
}

// dynamicHostPolicy only allows domains that exist in the DB
func dynamicHostPolicy(ctx context.Context, host string) error {
	var route db.Route
	result := db.DB.Where("domain = ?", host).First(&route)
	if result.Error != nil {
		log.Printf("ACME host validation failed for %s: %v", host, result.Error)
		return fmt.Errorf("unauthorized domain: %s", host)
	}
	return nil
}
