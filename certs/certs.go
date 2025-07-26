package certs

import (
	"crypto/tls"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var certCache = make(map[string]*tls.Certificate)
var domain = "timsalokat.dev"
var certBasePath = "/etc/letsencrypt"

func RenewCerts() {
	go func() {
		for {
			log.Println("Starting cert renewal...")

			cmd := exec.Command("certbot", "renew", "--standalone")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Renewal error: %v\n%s", err, string(out))
			}
			time.Sleep(12 * time.Hour)
		}
	}()
}

func CreateCertificates() error {
	_, err := getCertificate(domain)
	if err != nil {
		err = createCertificate(domain)
		if err != nil {
			return err
		}
	}

	_, err = getCertificate("*" + domain)
	if err != nil {
		err = createCertificate("*" + domain)
	}
	return err
}

func GetCertificates() []tls.Certificate {
	baseCertificate, err := getCertificate(domain)
	if err != nil {
		log.Printf("Error obtaining certificate for base domain: %v", err)
		log.Panic("Couldnt get or obtain cert for base domain")
	}

	subCertificate, err := getCertificate("*" + domain)
	if err != nil {
		log.Printf("Error obtaining certificate for base domain: %v", err)
		log.Panic("Couldnt get or obtain cert for sub domain")
	}
	return []tls.Certificate{*baseCertificate, *subCertificate}
}

func getCertificate(domain string) (*tls.Certificate, error) {
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

func createCertificate(domain string) error {

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
		log.Printf("Certbot failed for %s: %v\nOutput: %s", domain, err, string(out))
		return err
	}

	log.Printf("Certbot succeeded for %s:\n%s", domain, string(out))
	return nil
}
