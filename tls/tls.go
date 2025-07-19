package tls

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/timsalokat/latios_proxy/handlers"
)

var certBasePath = "/app/certs"

// ListAvailableCerts returns a slice of domain names for which certificates are available
func ListAvailableCerts() ([]string, error) {
	entries, err := os.ReadDir(certBasePath)
	if err != nil {
		println(fmt.Sprintf("Error reading cert directory: %v", err))
		return nil, err
	}

	var domains []string
	for _, entry := range entries {
		if entry.IsDir() {
			certPath := filepath.Join(certBasePath, entry.Name(), "fullchain.pem")
			keyPath := filepath.Join(certBasePath, entry.Name(), "privkey.pem")
			if _, err1 := os.Stat(certPath); err1 == nil {
				if _, err2 := os.Stat(keyPath); err2 == nil {
					domains = append(domains, entry.Name())
					println(fmt.Sprintf("Found valid cert for domain: %s", entry.Name()))
				} else {
					println(fmt.Sprintf("Missing key for domain: %s", entry.Name()))
				}
			} else {
				println(fmt.Sprintf("Missing cert for domain: %s", entry.Name()))
			}
		}
	}

	println(fmt.Sprintf("Available certs: %v", domains))

	return domains, nil
}

// getCertificateFunc dynamically loads a certificate for the given domain
func getCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	certCache := make(map[string]*tls.Certificate)

	println("getting certificates")
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
			println(fmt.Sprintf("Certificate not found for domain: %s", domain))
			return nil, err
		}
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			println(fmt.Sprintf("Key not found for domain: %s", domain))
			return nil, err
		}

		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			println(fmt.Sprintf("Failed loading cert for %s: %v", domain, err))
			return nil, err
		}

		println(fmt.Sprintf("Loaded cert for domain: %s", domain))
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
	if r.URL.Path == "/api/routes" {
		println("test")
		ListAvailableCerts()
		handlers.HandleRoutes(w, r)
		return
	}
	host := strings.Split(r.Host, ":")[0]
	target := "https://" + host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
