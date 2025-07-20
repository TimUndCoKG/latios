package db

type Route struct {
	ID              uint   `gorm:"primaryKey"`
	Domain          string `gorm:"uniqueIndex"`
	Target          string
	UseHTTPS        bool
	Static          bool
	StaticPath      string
	AllowWebsockets bool
}
