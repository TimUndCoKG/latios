package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Client *gorm.DB
var MemoryRoutes = make(map[string]Route)

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "latios-db"),
		getEnv("DB_USER", "user"),
		getEnv("DB_PASSWORD", "pass"),
		getEnv("DB_NAME", "latios"),
		getEnv("DB_PORT", "5432"),
	)

	var err error
	Client, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = Client.AutoMigrate(&Route{})
	if err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	err = loadRoutesIntoMemory()
	if err != nil {
		log.Fatalf("failed to load routes into memory: %v", err)
	}
}

func loadRoutesIntoMemory() error {
	var RouteList []Route
	result := Client.Find(&RouteList)
	if result.Error != nil {
		log.Fatalf("failed to load routes into memory: %v", result.Error)
		return result.Error
	}
	for _, route := range RouteList {
		MemoryRoutes[route.Domain] = route
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func GetRoute(domain string) (Route, error) {
	route := MemoryRoutes[domain]
	if route == (Route{}) {
		log.Print("Route not in memory, checking database")
		if Client.Where("domain = ?", domain).First(&route).Error != nil {
			return route, fmt.Errorf("route not found")
		}

		MemoryRoutes[domain] = route
	}
	return route, nil
}
