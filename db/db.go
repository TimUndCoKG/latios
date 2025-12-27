package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var Client *gorm.DB
var MemoryRoutes = make(map[string]Route)
var routeCacheLock sync.RWMutex

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "latios-db"),
		getEnv("DB_USER", "user"),
		getEnv("DB_PASSWORD", "pass"),
		getEnv("DB_NAME", "latios"),
		getEnv("DB_PORT", "5432"),
	)

	var err error
	gormConfig := &gorm.Config{
		Logger: gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormLogger.Config{
				LogLevel: gormLogger.Silent,
			},
		),
	}
	Client, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = Client.AutoMigrate(
		&Route{},
		&RequestLog{},
		&User{},
	)
	if err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	err = ensureBaseUser()
	if err != nil {
		log.Fatalf("failed to ensure base admin user: %v", err)
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

	routeCacheLock.Lock()
	defer routeCacheLock.Unlock()

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

func ensureBaseUser() error {
	var userCount int64
	err := Client.Model(&User{}).Count(&userCount).Error
	if err != nil {
		return err
	}

	if userCount == 0 {
		log.Printf("[AUTH] Creating admin user")

		bytes := make([]byte, 16)
		rand.Read(bytes)
		password := base64.StdEncoding.EncodeToString(bytes)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
			return err
		}

		user := User{
			Username: "admin",
			Password: string(hashedPassword),
		}

		if err := Client.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create base user: %v", err)
			return err
		}

		log.Println("######################################")
		log.Printf("[AUTH] Base user created")
		log.Printf("[AUTH] Username: admin")
		log.Printf("[AUTH] Password: %s", password)
		log.Println("######################################")
	}

	return nil

}

func GetRoute(domain string) (Route, error) {

	// * This can be put beforehand to optimize cache access
	// routeCacheLock.RLock()
	// route, ok := MemoryRoutes[domain]
	// routeCacheLock.RUnlock()

	// if ok {
	// 	return route, nil
	// }

	routeCacheLock.Lock()
	defer routeCacheLock.Unlock()

	route, ok := MemoryRoutes[domain]
	if ok {
		return route, nil
	}

	log.Print("Route not in memory, checking database")
	if Client.Where("domain = ?", domain).First(&route).Error != nil {
		return route, fmt.Errorf("route not found")
	}

	MemoryRoutes[domain] = route
	return route, nil
}

func AddRouteToCache(route Route) {
	routeCacheLock.Lock()
	defer routeCacheLock.Unlock()
	MemoryRoutes[route.Domain] = route
	log.Printf("[CACHE] Added route: %s", route.Domain)
}

func DeleteRouteFromCache(domain string) {
	routeCacheLock.Lock()
	defer routeCacheLock.Unlock()
	delete(MemoryRoutes, domain)
	log.Printf("[CACHE] Deleted route: %s", domain)
}
