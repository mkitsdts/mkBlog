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
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	jsonDecoder := json.NewDecoder(file)
	var config Config
	err = jsonDecoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
