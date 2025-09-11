package pkg

import (
	"mkBlog/config"
	"mkBlog/models"
	"os"
	"path"
)

func SaveImage(img *models.Image) error {
	filePath := path.Join(config.Cfg.Server.ImageSavePath, img.Name)
	return os.WriteFile(filePath, img.Data, 0644)
}
