package config

import "mkBlog/models"

func useDefaultConfig() {
	Cfg.Database.Kind = models.SQLite3

	Cfg.Server.Port = models.Default_Server_Port
	Cfg.Server.ImageSavePath = models.Default_Image_Save_Path
	Cfg.Server.Limiter.Duration = models.Default_Limiter_Duartion
	Cfg.Server.Limiter.Requests = models.Default_Limiter_Requests
	Cfg.Server.HTTP3Enabled = false
	Cfg.Server.Devmode = false

	Cfg.TLS.Enabled = false

	Cfg.CertControl.Enabled = false
}
