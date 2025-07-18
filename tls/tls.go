package tls

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func ServeWithTLS(handler http.Handler) {
	m := &autocert.Manager{
		Cache:      autocert.DirCache(".certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(), // Dynamic loading could be added later
	}

	httpsSrv := &http.Server{
		Addr:      ":443",
		Handler:   handler,
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}

	// Redirect HTTP to HTTPS
	go func() {
		log.Fatal(http.ListenAndServe(":80", m.HTTPHandler(nil)))
	}()

	log.Println("Listening on :443 with TLS")
	log.Fatal(httpsSrv.ListenAndServeTLS("", ""))
}
