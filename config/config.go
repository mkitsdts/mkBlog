package config

import (
	"encoding/json"
	"os"
)

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Config struct {
	MySQL MySQLConfig `json:"mysql"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	// Prefer environment variables
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.MySQL.Host = host
		cfg.MySQL.Port = os.Getenv("DB_PORT")
		cfg.MySQL.User = os.Getenv("DB_USER")
		cfg.MySQL.Password = os.Getenv("DB_PASSWORD")
		cfg.MySQL.Name = os.Getenv("DB_NAME")
		return &cfg, nil
	}

	// Fallback to config.json file if exists
	file, err := os.Open("config.json")
	if err != nil {
		return &cfg, nil // silent fallback (empty values)
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	if err := dec.Decode(&cfg); err != nil {
		return &cfg, nil
	}
	return &cfg, nil
}
