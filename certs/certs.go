package certs

import (
	"crypto/tls"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/timsalokat/latios_proxy/config"
)

var certCache = make(map[string]*tls.Certificate)
var certBasePath = "/etc/letsencrypt/live"

func RenewCerts() {
	go func() {
		for {
			log.Println("Starting cert renewal...")

			credFile, err := createCredentialFile()
			if err != nil {
				log.Fatal("Couldnt get cloudflare credential")
			}

			cmd := exec.Command("certbot", "renew",
				"--dns-cloudflare",
				"--dns-cloudflare-credentials", credFile.Name(),
				"--quiet")

			credFile.Close()
			os.Remove(credFile.Name())

			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Renewal error: %v\n%s", err, string(out))
			} else {
				log.Printf("Renewal output: %s", string(out))
			}

			time.Sleep(12 * time.Hour)
		}
	}()
}

func CreateCertificates() error {
	_, err := getCertificate(config.GetDomain())
	_, err2 := getCertificate(config.GetWildcardDomain())
	if err != nil || err2 != nil {
		err = createCertificate(config.GetDomain())
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCertificates() []tls.Certificate {
	baseCertificate, err := getCertificate(config.GetDomain())
	if err != nil {
		log.Printf("Error obtaining certificate for base domain: %v", err)
		log.Panic("Couldnt get or obtain cert for base domain")
	}

	subCertificate, err := getCertificate(config.GetWildcardDomain())
	if err != nil {
		log.Printf("Error obtaining certificate for wildcard domain: %v", err)
		log.Panic("Couldnt get or obtain cert for wildcard domain")
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
	credFile, err := createCredentialFile()
	if err != nil {
		log.Fatal("Couldnt get cloudflare credential")
	}
	defer credFile.Close()
	defer os.Remove(credFile.Name())

	cmd := exec.Command(
		"certbot",
		"certonly",
		"--dns-cloudflare",
		"--dns-cloudflare-credentials", credFile.Name(),
		"-d", domain,
		"-d", config.GetWildcardDomain(),
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

func createCredentialFile() (*os.File, error) {
	token := os.Getenv("CF_API_TOKEN")
	if token == "" {
		log.Fatal("CF_API_TOKEN not set")
	}

	credFile, err := os.CreateTemp("", "cloudflare.ini")
	if err != nil {
		log.Fatal(err)
	}

	_, err = credFile.WriteString("dns_cloudflare_api_token = " + token + "/n")
	if err != nil {
		log.Fatal(err)
	}

	return credFile, nil
}
