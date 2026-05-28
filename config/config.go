package config

import (
	"os"
)

type Config struct {
	ServerPort string
	DBPath     string
	UploadDir  string
	JWTSecret  string
	AdminUser  string
	AdminPass  string
	MaxImgSize int64 // bytes
	ImgMaxWidth int   // pixels
	ImgQuality  int   // JPEG quality 1-100
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("PORT", "8080"),
		DBPath:      getEnv("DB_PATH", "data/memorial.db"),
		UploadDir:   getEnv("UPLOAD_DIR", "uploads"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		AdminUser:   getEnv("ADMIN_USER", "admin"),
		AdminPass:   getEnv("ADMIN_PASS", "admin123"),
		MaxImgSize:  2 << 20, // 2MB
		ImgMaxWidth: 800,
		ImgQuality:  75,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
