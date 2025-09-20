package pkg

import (
	"encoding/base64"
	"mkBlog/config"
	"mkBlog/models"
	"os"
	"path"
)

func SaveImage(img *models.Image) error {
	filePath := path.Join(config.Cfg.Server.ImageSavePath, img.Name)
	data, err := base64.StdEncoding.DecodeString(img.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
