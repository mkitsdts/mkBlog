package config

import (
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
	BgPicturePath  string `json:"bgPicturePath" yaml:"bgPicturePath"`
	Server         string `json:"server" yaml:"server"`
	DevMode        bool   `json:"devmode" yaml:"devmode"`
	CommentEnabled bool   `json:"comment_enabled" yaml:"comment_enabled"`
	ICP            string `json:"icp" yaml:"icp"`
}

type ServerConfig struct {
	Port           int    `json:"port" yaml:"port"`
	Host           string `json:"host" yaml:"host"`
	DataPath       string `json:"-" yaml:"-"`
	LogFilePath    string `json:"-" yaml:"-"`
	ImageSavePath  string `json:"imageSavePath" yaml:"imageSavePath"`
	ConfigFilePath string `json:"-" yaml:"-"`
	DataFilePath   string `json:"-" yaml:"-"`
	StaticFilePath string `json:"-" yaml:"-"`
	Limiter        struct {
		Requests int `json:"requests" yaml:"requests"`
		Duration int `json:"duration" yaml:"duration"`
	} `json:"limiter" yaml:"limiter"`
	Devmode      bool `json:"devmode" yaml:"devmode"`
	HTTP3Enabled bool `json:"http3_enabled" yaml:"http3_enabled"`
}

type Config struct {
	Server      ServerConfig             `json:"server" yaml:"server"`
	Database    DatabaseConfig           `json:"database" yaml:"database"`
	TLS         TLSConfig                `json:"tls" yaml:"tls"`
	CertControl TLSCertAutoControlConfig `json:"cert_control" yaml:"cert_control"`
	Auth        AuthConfig               `json:"auth" yaml:"auth"`
	Site        SiteConfig               `json:"site" yaml:"site"`
}

func defaultServerConfig() ServerConfig {
	cfg := ServerConfig{
		Port:           models.Default_Server_Port,
		DataPath:       models.Default_Data_Path,
		LogFilePath:    models.Default_Log_File_Path,
		ImageSavePath:  models.Default_Image_Save_Path,
		ConfigFilePath: models.Default_Config_File_Path,
		DataFilePath:   models.Default_Data_File_Path,
		StaticFilePath: models.Default_Static_File_Path,
	}
	cfg.Limiter.Duration = models.Default_Limiter_Duartion
	cfg.Limiter.Requests = models.Default_Limiter_Requests
	return cfg
}

var Cfg = &Config{
	Server: defaultServerConfig(),
}

func Init() {
	// Fallback to config.yaml file if exists
	configPath := path.Join(Cfg.Server.DataPath, Cfg.Server.ConfigFilePath)
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
}

func Finalize() {
	if Cfg.Site.Server != "" {
		if normalized, err := normalizeServerURL(Cfg.Site.Server, Cfg.TLS.Enabled, Cfg.Server.Port); err != nil {
			slog.Warn("invalid site.server, using raw value", "server", Cfg.Site.Server, "error", err)
		} else {
			Cfg.Site.Server = normalized
		}
	}
	Cfg.Site.DevMode = Cfg.Server.Devmode

	if Cfg.TLS.Enabled {
		Cfg.TLS.Cert = path.Join(Cfg.Server.DataPath, Cfg.TLS.Cert)
		Cfg.TLS.Key = path.Join(Cfg.Server.DataPath, Cfg.TLS.Key)
	}

	slog.Info("Configuration loaded", "database", Cfg.Database, "tls", Cfg.TLS, "auth_enabled", Cfg.Auth.Enabled, "server", Cfg.Server)
}
