package certs

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/cloudflare"
	"github.com/timsalokat/latios_proxy/config"
)

func SetupTLSConfig() *tls.Config {
	domain := config.GetDomain()
	token := os.Getenv("CF_API_TOKEN")

	if token == "" {
		log.Fatal("[CERTS] Cloudflare API token not set")
	}

	log.Println("[CERTS] Configuring certmagic with Cloudflare...")

	certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
		DNSManager: certmagic.DNSManager{
			DNSProvider: &cloudflare.Provider{
				APIToken: token,
			},
		},
	}

	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.Email = "admin@" + domain

	// certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA

	cfg := certmagic.NewDefault()
	err := cfg.ManageSync(context.Background(), []string{
		domain,
		"*." + domain,
	})
	if err != nil {
		log.Fatalf("[CERTS] Failed to manage certificates: %v", err)
	}

	return cfg.TLSConfig()
}
