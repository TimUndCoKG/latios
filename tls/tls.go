package tls

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/timsalokat/latios_proxy/handlers"
)

var certBasePath = "/app/certs"

// getLatestCertFiles finds the highest numbered fullchain/privkey pair in the archive dir
func getLatestCertFiles(domain string) (string, string, error) {
	domainPath := filepath.Join(certBasePath, domain)
	entries, err := os.ReadDir(domainPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read domain cert folder: %w", err)
	}

	var versions []int
	re := regexp.MustCompile(`^fullchain(\d+)\.pem$`)

	for _, entry := range entries {
		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) == 2 {
			num, err := strconv.Atoi(matches[1])
			if err == nil {
				versions = append(versions, num)
			}
		}
	}

	if len(versions) == 0 {
		return "", "", fmt.Errorf("no fullchainN.pem found for domain %s", domain)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(versions)))
	n := versions[0]

	fullchain := filepath.Join(domainPath, fmt.Sprintf("fullchain%d.pem", n))
	privkey := filepath.Join(domainPath, fmt.Sprintf("privkey%d.pem", n))

	return fullchain, privkey, nil
}

// getCertificateFunc dynamically loads a certificate for the given domain
func getCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	certCache := make(map[string]*tls.Certificate)

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName
		if domain == "" {
			return nil, fmt.Errorf("no SNI server name")
		}

		// Check cache
		if cert, ok := certCache[domain]; ok {
			return cert, nil
		}

		fullchain, privkey, err := getLatestCertFiles(domain)
		if err != nil {
			println(fmt.Sprintf("No certs found for domain %s: %v", domain, err))
			return nil, err
		}

		cert, err := tls.LoadX509KeyPair(fullchain, privkey)
		if err != nil {
			return nil, fmt.Errorf("failed to load cert for %s: %w", domain, err)
		}

		println(fmt.Sprintf("Loaded cert vN for domain: %s", domain))
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

	go func() {
		log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(redirectToHTTPS)))
	}()

	log.Println("Listening on :443 with dynamic TLS certs")
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/routes" {
		handlers.HandleRoutes(w, r)
		return
	}
	host := strings.Split(r.Host, ":")[0]
	target := "https://" + host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
