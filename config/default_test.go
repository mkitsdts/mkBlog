package config

import (
	"mkBlog/models"
	"path"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestDefaultServerConfig(t *testing.T) {
	cfg := defaultServerConfig()

	if cfg.DataPath != models.Default_Data_Path {
		t.Errorf("data path = %q, want %q", cfg.DataPath, models.Default_Data_Path)
	}
	if cfg.LogFilePath != models.Default_Log_File_Path {
		t.Errorf("log file path = %q, want %q", cfg.LogFilePath, models.Default_Log_File_Path)
	}
	if cfg.ImageSavePath != models.Default_Image_Save_Path {
		t.Errorf("image save path = %q, want %q", cfg.ImageSavePath, models.Default_Image_Save_Path)
	}
	if cfg.ConfigFilePath != models.Default_Config_File_Path {
		t.Errorf("config file path = %q, want %q", cfg.ConfigFilePath, models.Default_Config_File_Path)
	}
	if cfg.DataFilePath != models.Default_Data_File_Path {
		t.Errorf("data file path = %q, want %q", cfg.DataFilePath, models.Default_Data_File_Path)
	}
	if cfg.Port != models.Default_Server_Port {
		t.Errorf("server port = %d, want %d", cfg.Port, models.Default_Server_Port)
	}
	if cfg.Limiter.Duration != models.Default_Limiter_Duartion {
		t.Errorf("limiter duration = %d, want %d", cfg.Limiter.Duration, models.Default_Limiter_Duartion)
	}
	if cfg.Limiter.Requests != models.Default_Limiter_Requests {
		t.Errorf("limiter requests = %d, want %d", cfg.Limiter.Requests, models.Default_Limiter_Requests)
	}
}

func TestRuntimePathsSurviveConfigDecode(t *testing.T) {
	cfg := Config{Server: defaultServerConfig()}
	input := []byte("server:\n  port: 9000\n  imageSavePath: custom-images\n")

	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("decode config: %v", err)
	}

	if cfg.Server.DataPath != models.Default_Data_Path {
		t.Errorf("data path changed to %q", cfg.Server.DataPath)
	}
	if cfg.Server.LogFilePath != models.Default_Log_File_Path {
		t.Errorf("log file path changed to %q", cfg.Server.LogFilePath)
	}
	if cfg.Server.ConfigFilePath != models.Default_Config_File_Path {
		t.Errorf("config file path changed to %q", cfg.Server.ConfigFilePath)
	}
	if cfg.Server.DataFilePath != models.Default_Data_File_Path {
		t.Errorf("data file path changed to %q", cfg.Server.DataFilePath)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("server port = %d, want 9000", cfg.Server.Port)
	}
	if cfg.Server.ImageSavePath != "custom-images" {
		t.Errorf("image save path = %q, want custom-images", cfg.Server.ImageSavePath)
	}
}

func TestUseDefaultConfigDataPath(t *testing.T) {
	originalCfg := Cfg
	t.Cleanup(func() {
		Cfg = originalCfg
	})

	tests := []struct {
		name     string
		dataPath string
		want     string
	}{
		{
			name:     "absolute path",
			dataPath: "/etc/mkblog",
			want:     "/etc/mkblog/img",
		},
		{
			name:     "relative path",
			dataPath: "./custom-data",
			want:     path.Join(PWD(), "custom-data", "img"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Cfg = &Config{
				Server: ServerConfig{
					DataPath: tt.dataPath,
				},
			}

			useDefaultConfig()

			if got := Cfg.Server.ImageSavePath; got != tt.want {
				t.Fatalf("image save path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFinalizeImageSavePath(t *testing.T) {
	originalCfg := Cfg
	t.Cleanup(func() {
		Cfg = originalCfg
	})

	tests := []struct {
		name          string
		dataPath      string
		imageSavePath string
		want          string
	}{
		{
			name:          "new default",
			dataPath:      "./data",
			imageSavePath: "img",
			want:          filepath.Join("data", "img"),
		},
		{
			name:          "migrate legacy generated default",
			dataPath:      "./data",
			imageSavePath: "static/images",
			want:          filepath.Join("data", "img"),
		},
		{
			name:          "custom relative directory",
			dataPath:      "./data",
			imageSavePath: "pictures",
			want:          filepath.Join("data", "pictures"),
		},
		{
			name:          "custom absolute directory",
			dataPath:      "./data",
			imageSavePath: "/srv/mkblog-pictures",
			want:          "/srv/mkblog-pictures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Cfg = &Config{
				Server: ServerConfig{
					DataPath:      tt.dataPath,
					ImageSavePath: tt.imageSavePath,
				},
			}

			Finalize()

			if got := Cfg.Server.ImageSavePath; got != tt.want {
				t.Fatalf("image save path = %q, want %q", got, tt.want)
			}
		})
	}
}
