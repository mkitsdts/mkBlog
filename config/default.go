package config

import (
	"mkBlog/models"
	"path"
)

func useDefaultConfig() {
	pwd := PWD()
	dataPath := path.Join(pwd, models.Default_Data_Path)
	Cfg.Database.Kind = models.SQLite3

	Cfg.Server.Port = models.Default_Server_Port
	Cfg.Server.ImageSavePath = path.Join(dataPath, models.Default_Image_Save_Path)
	Cfg.Server.Limiter.Duration = models.Default_Limiter_Duartion
	Cfg.Server.Limiter.Requests = models.Default_Limiter_Requests
	Cfg.Server.HTTP3Enabled = false
	Cfg.Server.Devmode = false

	Cfg.TLS.Enabled = false

	Cfg.CertControl.Enabled = false
}
