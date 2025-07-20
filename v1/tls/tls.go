package tls

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/timsalokat/latios_proxy/handlers"
)

var (
	// certBasePath = "/app/certs"
	certBasePath = "/etc/letsencrypt"
	httpServer   *http.Server
	mu           sync.Mutex
	certCache    = make(map[string]*tls.Certificate)
)

// ListAvailableCerts returns a slice of domain names with valid certs
func ListAvailableCerts() ([]string, error) {
	entries, err := os.ReadDir(certBasePath)
	if err != nil {
		log.Printf("Error reading cert directory: %v", err)
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
					log.Printf("Found valid cert for domain: %s", entry.Name())
				}
			}
		}
	}
	return domains, nil
}

// GetCertificateFunc loads certs dynamically based on SNI
func GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName
		if domain == "" {
			return nil, fmt.Errorf("no SNI server name")
		}

		mu.Lock()
		defer mu.Unlock()

		if cert, ok := certCache[domain]; ok {
			return cert, nil
		}

		certPath := filepath.Join(certBasePath, domain, "fullchain.pem")
		keyPath := filepath.Join(certBasePath, domain, "privkey.pem")

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			log.Printf("Cert not found for domain: %s", domain)
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

// RunCertbot issues a cert for the given domain using --standalone
func RunCertbot(domain string) error {
	log.Printf("Issuing cert for %s...", domain)
	cmd := exec.Command(
		"certbot",
		"certonly",
		"--standalone",
		"-d", domain,
		"--agree-tos",
		"--no-eff-email",
		"-m", "admin@"+domain,
		"--non-interactive",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Certbot failed: %v\n%s", err, string(out))
		return err
	}
	log.Printf("Certbot output:\n%s", string(out))
	return nil
}

// RenewCerts periodically renews all existing certs
func RenewCerts() {
	go func() {
		for {
			log.Println("Starting cert renewal...")
			StopHTTPServer()

			cmd := exec.Command("certbot", "renew", "--standalone")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Renewal error: %v\n%s", err, string(out))
			} else {
				log.Printf("Renewal succeeded:\n%s", string(out))
				// Clear the cache to reload renewed certs
				mu.Lock()
				certCache = make(map[string]*tls.Certificate)
				mu.Unlock()
			}

			StartHTTPRedirect()
			time.Sleep(12 * time.Hour)
		}
	}()
}

func StopHTTPServer() {
	if httpServer != nil {
		log.Println("Stopping HTTP redirect server for Certbot...")
		_ = httpServer.Close()
	}
}

func StartHTTPRedirect() {
	go func() {
		httpServer = &http.Server{
			Addr: ":80",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/routes" {
					handlers.HandleRoutes(w, r)
					return
				}
				host := strings.Split(r.Host, ":")[0]
				target := "https://" + host + r.URL.RequestURI()
				http.Redirect(w, r, target, http.StatusMovedPermanently)
			}),
		}
		log.Println("Starting HTTP redirect server on port 80")
		log.Fatal(httpServer.ListenAndServe())
	}()
}

// ServeWithTLS launches the HTTPS server with dynamic cert loading
func ServeWithTLS(handler http.Handler) {
	// Start the redirect server first
	StartHTTPRedirect()

	RenewCerts() // start background renewer

	server := &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: GetCertificateFunc(),
			MinVersion:     tls.VersionTLS12,
		},
	}

	log.Println("Listening on :443 with manual per-domain TLS")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
