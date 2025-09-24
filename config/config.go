package config

import (
	"log/slog"
	"os"

	"go.yaml.in/yaml/v3"
)

type MySQLConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
}

type TLSConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Cert    string `json:"cert" yaml:"cert"`
	Key     string `json:"key" yaml:"key"`
}

type AuthConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Secret  string `json:"secret" yaml:"secret"`
}

type ServerConfig struct {
	Port          int    `json:"port" yaml:"port"`
	Host          string `json:"host" yaml:"host"`
	ImageSavePath string `json:"imageSavePath" yaml:"imageSavePath"`
	Limiter       struct {
		Requests int `json:"requests" yaml:"requests"`
		Duration int `json:"duration" yaml:"duration"`
	} `json:"limiter" yaml:"limiter"`
	Devmode bool `json:"devmode" yaml:"devmode"`
}

type Config struct {
	Server ServerConfig `json:"server" yaml:"server"`
	MySQL  MySQLConfig  `json:"mysql" yaml:"mysql"`
	TLS    TLSConfig    `json:"tls" yaml:"tls"`
	Auth   AuthConfig   `json:"auth" yaml:"auth"`
}

var Cfg *Config = &Config{}

func init() {
	// Fallback to config.yaml file if exists
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Warn("config.yaml not found, using environment variables or defaults")
	} else {
		defer file.Close()
		if err := yaml.NewDecoder(file).Decode(Cfg); err != nil {
			slog.Warn("Failed to decode config.yaml, using environment variables or defaults")
		}
		slog.Info("Configuration loaded", "mysql", Cfg.MySQL, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled)
		return // Loaded from file, skip env vars
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		Cfg.MySQL.Host = host
	} else {
		Cfg.MySQL.Host = "localhost"
		slog.Warn("DB host not set, defaulting to localhost")
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		Cfg.MySQL.Port = port
	} else {
		Cfg.MySQL.Port = "3306"
		slog.Warn("DB port not set, defaulting to 3306")
	}

	if user := os.Getenv("DB_USER"); user != "" {
		Cfg.MySQL.User = user
	} else {
		Cfg.MySQL.User = "root"
		slog.Warn("DB user not set, defaulting to root")
	}

	if password := os.Getenv("DB_PASSWORD"); password != "" {
		Cfg.MySQL.Password = password
	} else {
		Cfg.MySQL.Password = "root"
		slog.Warn("DB password not set, defaulting to root")
	}

	if name := os.Getenv("DB_NAME"); name != "" {
		Cfg.MySQL.Name = name
	} else {
		Cfg.MySQL.Name = "mkblog"
		slog.Warn("DB name not set, defaulting to mkblog")
	}

	if auth := os.Getenv("AUTH_ENABLE"); auth != "" {
		Cfg.Auth.Enabled = auth == "true" || auth == "1"
	}

	if Cfg.Auth.Enabled {
		if secret := os.Getenv("AUTH_SECRET"); secret != "" {
			Cfg.Auth.Secret = secret
		} else {
			Cfg.Auth.Secret = ""
			slog.Warn("Auth secret not set, using default (insecure)")
		}
	}

	if tls := os.Getenv("TLS_ENABLE"); tls != "" {
		Cfg.TLS.Enabled = tls == "true" || tls == "1"
	}

	if !Cfg.TLS.Enabled {
		slog.Info("loaded configuration successfully", "mysql", Cfg.MySQL, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled)
		slog.Warn("TLS is disabled, consider enabling it in production environments")
		panic("TLS is disabled, please enable it for better security")
	}

	if cert := os.Getenv("TLS_CERT"); cert != "" {
		Cfg.TLS.Cert = cert
	} else {
		Cfg.TLS.Cert = "localhost.crt"
		slog.Warn("TLS cert not set, defaulting to localhost.crt")
	}

	if key := os.Getenv("TLS_KEY"); key != "" {
		Cfg.TLS.Key = key
	} else {
		Cfg.TLS.Key = "localhost.key"
		slog.Warn("TLS key not set, defaulting to localhost.key")
	}

	slog.Info("Configuration loaded", "mysql", Cfg.MySQL, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled)

}
