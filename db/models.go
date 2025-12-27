package db

import "time"

type Route struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Domain     string `gorm:"uniqueIndex" json:"domain"`
	TargetPath string `json:"target_path"`
	// UseHTTPS    bool   `json:"use_https"`
	IsStatic    bool `json:"is_static"`
	EnforceAuth bool `json:"enforce_auth"`
}

type RequestLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Timestamp  time.Time `gorm:"index" json:"timestamp"`
	Method     string    `json:"method"`
	Host       string    `gorm:"index" json:"host"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	LatencyMs  int64     `json:"latency_ms"`
	RemoteAddr string    `json:"remote_addr"`
}
