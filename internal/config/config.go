package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	JWT    JWTConfig
	Upload UploadConfig
}

type ServerConfig struct {
	Port string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	ResidentSecret string
	AdminSecret    string
	ExpireHours    int
}

type UploadConfig struct {
	LocalPath  string
	BaseURL    string
	MaxSize    int64
	MinCount   int
	MaxCount   int
	AllowedExt string
}

var Cfg *Config

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "123456"),
			DBName:   getEnv("MYSQL_DBNAME", "appliance_recycle"),
		},
		JWT: JWTConfig{
			ResidentSecret: getEnv("JWT_RESIDENT_SECRET", "resident_secret_change_me"),
			AdminSecret:    getEnv("JWT_ADMIN_SECRET", "admin_secret_change_me"),
			ExpireHours:    parseInt(getEnv("JWT_EXPIRE_HOURS", "72"), 72),
		},
		Upload: UploadConfig{
			LocalPath:  getEnv("UPLOAD_LOCAL_PATH", "./uploads"),
			BaseURL:    getEnv("UPLOAD_BASE_URL", "http://localhost:8080/uploads"),
			MaxSize:    int64(parseInt(getEnv("UPLOAD_MAX_SIZE_MB", "5"), 5)) * 1024 * 1024,
			MinCount:   parseInt(getEnv("UPLOAD_MIN_COUNT", "3"), 3),
			MaxCount:   parseInt(getEnv("UPLOAD_MAX_COUNT", "5"), 5),
			AllowedExt: getEnv("UPLOAD_ALLOWED_EXT", ".jpg,.jpeg,.png,.webp"),
		},
	}

	Cfg = cfg
	return cfg, nil
}

func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DBName)
}

func (u UploadConfig) ExtSet() map[string]bool {
	m := make(map[string]bool)
	for _, e := range strings.Split(u.AllowedExt, ",") {
		p := strings.TrimSpace(e)
		if p != "" {
			m[p] = true
		}
	}
	return m
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseInt(s string, defaultValue int) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
