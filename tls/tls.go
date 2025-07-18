package tls

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var certBasePath = "/certs"

// getCertificateFunc dynamically loads a certificate for the given domain
func getCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	certCache := make(map[string]*tls.Certificate)

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName
		if domain == "" {
			return nil, fmt.Errorf("no SNI server name")
		}

		// Check cache first
		if cert, ok := certCache[domain]; ok {
			return cert, nil
		}

		certPath := filepath.Join(certBasePath, domain, "fullchain.pem")
		keyPath := filepath.Join(certBasePath, domain, "privkey.pem")

		// Verify both files exist
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			log.Printf("Certificate not found for domain: %s", domain)
			return nil, err
		}
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			log.Printf("Key not found for domain: %s", domain)
			return nil, err
		}

		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			log.Printf("Failed loading cert for %s: %v", domain, err)
			return nil, err
		}

		log.Printf("Loaded cert for domain: %s", domain)
		certCache[domain] = &cert
		return &cert, nil
	}
}

func ServeWithTLS(handler http.Handler) {
	server := &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: getCertificateFunc(),
			MinVersion:     tls.VersionTLS12,
		},
	}

	// Optionally redirect HTTP to HTTPS
	go func() {
		log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(redirectToHTTPS)))
	}()

	log.Println("Listening on :443 with manual per-domain TLS")
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/routes") {
		host := strings.Split(r.Host, ":")[0]
		target := "https://" + host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
}
