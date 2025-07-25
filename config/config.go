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
	var config Config
	file, err := os.Open("config.json")
	if err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&config); err != nil {
			return nil, err
		}
		return &config, err
	}
	config.MySQL.Host = os.Getenv("MYSQL_HOST")
	config.MySQL.Port = os.Getenv("MYSQL_PORT")
	config.MySQL.User = os.Getenv("MYSQL_USER")
	config.MySQL.Password = os.Getenv("MYSQL_PASSWORD")
	config.MySQL.Name = os.Getenv("MYSQL_NAME")
	return &config, nil
}
