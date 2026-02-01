package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	DSN      string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	DSN      string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type LogConfig struct {
	Level  string
	Format string
}

func getStringWithFallback(primaryKey, fallbackKey string) string {
	value := viper.GetString(primaryKey)
	if value == "" {
		value = viper.GetString(fallbackKey)
	}
	return value
}

func getStringSlice(key string) []string {
	str := viper.GetString(key)
	if str == "" {
		return viper.GetStringSlice(key)
	}
	// Split by comma and trim spaces
	var result []string
	for _, part := range strings.Split(str, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./..")
	viper.AddConfigPath("./../..")

	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("SERVER_READ_TIMEOUT", "10s")
	viper.SetDefault("SERVER_WRITE_TIMEOUT", "10s")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "massrouter")
	viper.SetDefault("DB_PASSWORD", "changeme")
	viper.SetDefault("DB_NAME", "massrouter")
	viper.SetDefault("DB_SSL_MODE", "disable")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "changeme")
	viper.SetDefault("REDIS_DB", "0")

	viper.SetDefault("JWT_SECRET", "your-secret-key-change-in-production")
	viper.SetDefault("JWT_ACCESS_EXPIRY", "24h")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	viper.SetDefault("JWT_ISSUER", "massrouter")

	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")
	viper.SetDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	viper.SetDefault("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization")
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", "true")

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getStringWithFallback("PORT", "SERVER_PORT"),
			Mode:         viper.GetString("SERVER_MODE"),
			ReadTimeout:  viper.GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout: viper.GetDuration("SERVER_WRITE_TIMEOUT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
			DSN:      buildDSN(viper.GetString("DB_HOST"), viper.GetString("DB_PORT"), viper.GetString("DB_USER"), viper.GetString("DB_PASSWORD"), viper.GetString("DB_NAME"), viper.GetString("DB_SSL_MODE")),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
			DSN:      buildRedisDSN(viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT"), viper.GetString("REDIS_PASSWORD"), viper.GetInt("REDIS_DB")),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  viper.GetDuration("JWT_ACCESS_EXPIRY"),
			RefreshExpiry: viper.GetDuration("JWT_REFRESH_EXPIRY"),
			Issuer:        viper.GetString("JWT_ISSUER"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getStringSlice("CORS_ALLOWED_ORIGINS"),
			AllowedMethods:   getStringSlice("CORS_ALLOWED_METHODS"),
			AllowedHeaders:   getStringSlice("CORS_ALLOWED_HEADERS"),
			AllowCredentials: viper.GetBool("CORS_ALLOW_CREDENTIALS"),
		},
		Log: LogConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
	}

	log.Printf("Server Port: %s", config.Server.Port)
	log.Printf("CORS Allowed Origins: %v", config.CORS.AllowedOrigins)
	log.Printf("CORS Allow Credentials: %v", config.CORS.AllowCredentials)

	return config, nil
}

func buildDSN(host, port, user, password, dbname, sslmode string) string {
	return "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
}

func buildRedisDSN(host, port, password string, db int) string {
	if password != "" {
		return "redis://" + password + "@" + host + ":" + port + "/" + string(rune(db))
	}
	return "redis://" + host + ":" + port + "/" + string(rune(db))
}
