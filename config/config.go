package config

import (
	"encoding/json"
	"log/slog"
	"os"
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

type Config struct {
	MySQL MySQLConfig `json:"mysql" yaml:"mysql"`
	TLS   TLSConfig   `json:"tls" yaml:"tls"`
	Auth  AuthConfig  `json:"auth" yaml:"auth"`
}

var Cfg *Config = &Config{}

func LoadConfig() error {
	// Fallback to config.json file if exists
	file, err := os.Open("config.json")
	if err != nil {
		slog.Warn("config.json not found, using environment variables or defaults")
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	if err := dec.Decode(&Cfg); err != nil {
		slog.Warn("Failed to decode config.json, using environment variables or defaults")
	}

	if host := os.Getenv("DB_HOST"); host != "" {
		Cfg.MySQL.Host = host
	} else if Cfg.MySQL.Host == "" {
		Cfg.MySQL.Host = "localhost"
		slog.Warn("DB host not set, defaulting to localhost")
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		Cfg.MySQL.Port = port
	} else if Cfg.MySQL.Port == "" {
		Cfg.MySQL.Port = "3306"
		slog.Warn("DB port not set, defaulting to 3306")
	}

	if user := os.Getenv("DB_USER"); user != "" {
		Cfg.MySQL.User = user
	} else if Cfg.MySQL.User == "" {
		Cfg.MySQL.User = "root"
		slog.Warn("DB user not set, defaulting to root")
	}

	if password := os.Getenv("DB_PASSWORD"); password != "" {
		Cfg.MySQL.Password = password
	} else if Cfg.MySQL.Password == "" {
		Cfg.MySQL.Password = "root"
		slog.Warn("DB password not set, defaulting to root")
	}

	if name := os.Getenv("DB_NAME"); name != "" {
		Cfg.MySQL.Name = name
	} else if Cfg.MySQL.Name == "" {
		Cfg.MySQL.Name = "mkblog"
		slog.Warn("DB name not set, defaulting to mkblog")
	}

	if tls := os.Getenv("TLS_ENABLE"); tls == "true" || tls == "1" {
		Cfg.TLS.Enabled = true
	} else {
		Cfg.TLS.Enabled = false
		slog.Warn("TLS not enabled, defaulting to false")
		return nil
	}

	if cert := os.Getenv("TLS_CERT"); cert != "" {
		Cfg.TLS.Cert = cert
	} else if Cfg.TLS.Cert == "" {
		Cfg.TLS.Cert = "localhost.crt"
		slog.Warn("TLS cert not set, defaulting to localhost.crt")
	}

	if key := os.Getenv("TLS_KEY"); key != "" {
		Cfg.TLS.Key = key
	} else if Cfg.TLS.Key == "" {
		Cfg.TLS.Key = "localhost.key"
		slog.Warn("TLS key not set, defaulting to localhost.key")
	}

	return nil
}
