package config

import (
	"fmt"
	"log/slog"
	"mkBlog/models"
	"os"
	"path"

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
	configPath := path.Join(models.Default_Data_Path, models.Default_Config_File_Path)
	file, err := os.Open(configPath)
	if err != nil {
		slog.Warn("config file not found, writing default config.yaml")
		if err = writeImpl(); err != nil {
			slog.Error("Failed to write impl config file.", " Please check program's permission ", err)
			useDefaultConfig()
			return
		}
		if file, err = os.Open(configPath); err != nil {
			slog.Error("Failed to open file.", " Unknown error: ", err)
			useDefaultConfig()
			return
		}
	}
	if err := yaml.NewDecoder(file).Decode(Cfg); err != nil {
		slog.Warn("Failed to decode config.yaml")
		return
	}

	Cfg.Site.Server = fmt.Sprintf("http://localhost:%d", Cfg.Server.Port)
	Cfg.Site.DevMode = Cfg.Server.Devmode

	if Cfg.TLS.Enabled {
		Cfg.TLS.Cert = path.Join(path.Join(models.Default_Data_Path, Cfg.TLS.Cert))
		Cfg.TLS.Key = path.Join(models.Default_Data_Path, Cfg.TLS.Key)
	}

	slog.Info("Configuration loaded", "database", Cfg.Database, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled, "server", Cfg.Server)
}
