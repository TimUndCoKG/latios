package certs

import (
	"crypto/tls"
	"log"
	"os/exec"
	"time"
)

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
