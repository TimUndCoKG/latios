package db

type Route struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Domain      string `gorm:"uniqueIndex" json:"domain"`
	TargetPath  string `json:"target_path"`
	UseHTTPS    bool   `json:"use_https"`
	IsStatic    bool   `json:"is_static"`
	EnforceAuth bool   `json:"enforce_auth"`
}
