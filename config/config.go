package config

import (
	"log"
	"os"
)

var DOMAIN string

func LoadConfig() {
	DOMAIN := os.Getenv("DOMAIN")
	if DOMAIN == "" {
		log.Fatal("DOMAIN not set")
		os.Exit(1)
	}
}

func GetDomain() string {
	return DOMAIN
}

func GetWildcardDomain() string {
	return "*." + DOMAIN
}
