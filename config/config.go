package config

import (
	"log/slog"
	"os"

	"go.yaml.in/yaml/v3"
)

type DatabaseConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
	Kind     string `json:"kind" yaml:"kind"`
}

type TLSConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Cert    string `json:"cert" yaml:"cert"`
	Key     string `json:"key" yaml:"key"`
}

type TLSCertAutoControlConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	Email          string `json:"email" yaml:"email"`
	Domain         string `json:"domain" yaml:"domain"`
	Key            string `json:"key" yaml:"key"`
	Secret         string `json:"secret" yaml:"secret"`
	DomainProvider string `json:"domain_provider" yaml:"domain_provider"`
}

type AuthConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Secret  string `json:"secret" yaml:"secret"`
}

type SiteConfig struct {
	Signature      string `json:"signature" yaml:"signature"`
	About          string `json:"about" yaml:"about"`
	AvatarPath     string `json:"avatarPath" yaml:"avatarPath"`
	Server         string `json:"server" yaml:"server"`
	DevMode        bool   `json:"devmode" yaml:"devmode"`
	CommentEnabled bool   `json:"comment_enabled" yaml:"comment_enabled"`
	ICP            string `json:"icp" yaml:"icp"`
}

type ServerConfig struct {
	Port          int    `json:"port" yaml:"port"`
	Host          string `json:"host" yaml:"host"`
	ImageSavePath string `json:"imageSavePath" yaml:"imageSavePath"`
	Limiter       struct {
		Requests int `json:"requests" yaml:"requests"`
		Duration int `json:"duration" yaml:"duration"`
	} `json:"limiter" yaml:"limiter"`
	Devmode                bool `json:"devmode" yaml:"devmode"`
	HTTP3Enabled           bool `json:"http3_enabled" yaml:"http3_enabled"`
	CertAutoControlEnabled bool `json:"cert_ctrl_enabled" yaml:"cert_ctrl_enabled"`
}

type Config struct {
	Server      ServerConfig             `json:"server" yaml:"server"`
	Database    DatabaseConfig           `json:"database" yaml:"database"`
	TLS         TLSConfig                `json:"tls" yaml:"tls"`
	CertControl TLSCertAutoControlConfig `json:"cert_control" yaml:"cert_control"`
	Auth        AuthConfig               `json:"auth" yaml:"auth"`
	Site        SiteConfig               `json:"site" yaml:"site"`
}

var Cfg *Config = &Config{}

func Init() {
	// Fallback to config.yaml file if exists
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Warn("config.yaml not found, using environment variables or defaults")
	} else {
		defer file.Close()
		if err := yaml.NewDecoder(file).Decode(Cfg); err != nil {
			slog.Warn("Failed to decode config.yaml, using environment variables or defaults")
		}
		slog.Info("Configuration loaded", "database", Cfg.Database, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled, "server", Cfg.Server)
		return // Loaded from file, skip env vars
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		Cfg.Database.Host = host
	} else {
		Cfg.Database.Host = "localhost"
		slog.Warn("DB host not set, defaulting to localhost")
	}

	if kind := os.Getenv("DB_KIND"); kind != "" {
		Cfg.Database.Kind = kind
	} else {
		Cfg.Database.Kind = "mysql"
		slog.Warn("DB kind not set, defaulting to mysql")
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		Cfg.Database.Port = port
	} else {
		switch Cfg.Database.Kind {
		case "postgres":
			Cfg.Database.Port = "5432"
			slog.Warn("DB port not set, defaulting to 5432 for postgres")
			return
		case "mysql":
			Cfg.Database.Port = "3306"
			slog.Warn("DB port not set, defaulting to 3306 for mysql")
			return
		default:
			Cfg.Database.Port = "3306"
			slog.Warn("DB port not set, defaulting to 3306")
		}
	}

	if user := os.Getenv("DB_USER"); user != "" {
		Cfg.Database.User = user
	} else {
		Cfg.Database.User = "root"
		slog.Warn("DB user not set, defaulting to root")
	}

	if password := os.Getenv("DB_PASSWORD"); password != "" {
		Cfg.Database.Password = password
	} else {
		Cfg.Database.Password = "root"
		slog.Warn("DB password not set, defaulting to root")
	}

	if name := os.Getenv("DB_NAME"); name != "" {
		Cfg.Database.Name = name
	} else {
		Cfg.Database.Name = "mkblog"
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
		slog.Info("loaded configuration successfully", "database", Cfg.Database, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled)
		slog.Warn("TLS is disabled, consider enabling it in production environments")
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

	slog.Info("Configuration loaded", "database", Cfg.Database, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled)

}
