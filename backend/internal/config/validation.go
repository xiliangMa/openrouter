package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() error {
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}
	if err := c.Database.validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}
	if err := c.Redis.validate(); err != nil {
		return fmt.Errorf("redis config: %w", err)
	}
	if err := c.JWT.validate(); err != nil {
		return fmt.Errorf("jwt config: %w", err)
	}
	return nil
}

func (c *ServerConfig) validate() error {
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	if c.Mode != "debug" && c.Mode != "release" && c.Mode != "test" {
		return fmt.Errorf("mode must be debug, release, or test")
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	return nil
}

func (c *DatabaseConfig) validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

func (c *RedisConfig) validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	return nil
}

func (c *JWTConfig) validate() error {
	if c.Secret == "" {
		return fmt.Errorf("secret is required")
	}
	if len(c.Secret) < 32 {
		return fmt.Errorf("secret must be at least 32 characters")
	}
	if c.AccessExpiry <= 0 {
		return fmt.Errorf("access expiry must be positive")
	}
	if c.RefreshExpiry <= 0 {
		return fmt.Errorf("refresh expiry must be positive")
	}
	if c.AccessExpiry >= c.RefreshExpiry {
		return fmt.Errorf("access expiry must be less than refresh expiry")
	}
	return nil
}

func (c *CORSConfig) validate() error {
	if len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("at least one allowed origin is required")
	}
	for _, origin := range c.AllowedOrigins {
		if strings.TrimSpace(origin) == "" {
			return fmt.Errorf("allowed origin cannot be empty")
		}
	}
	return nil
}

func (c *LogConfig) validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("log level must be debug, info, warn, error, or fatal")
	}
	if c.Format != "json" && c.Format != "text" {
		return fmt.Errorf("log format must be json or text")
	}
	return nil
}
