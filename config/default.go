package config

import (
	"mkBlog/models"
	"path"
)

func useDefaultConfig() {
	dataPath := Cfg.Server.DataPath
	if !path.IsAbs(dataPath) {
		dataPath = path.Join(PWD(), dataPath)
	}
	Cfg.Database.Kind = models.SQLite3

	defaults := defaultServerConfig()
	Cfg.Server.Port = defaults.Port
	Cfg.Server.ImageSavePath = path.Join(dataPath, defaults.ImageSavePath)
	Cfg.Server.Limiter = defaults.Limiter
	Cfg.Server.HTTP3Enabled = false
	Cfg.Server.Devmode = false

	Cfg.TLS.Enabled = false

	Cfg.CertControl.Enabled = false
}
