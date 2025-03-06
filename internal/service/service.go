package service

import (
	"mkBlog/internal/database"
	"mkBlog/utils/medicine"
	"path/filepath"
	"encoding/json"
	"os"
	"fmt"
)

// 写完文章后更新
func UpdateArticle() {
	err := filepath.Walk("resource", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if filepath.Ext(path) == ".md" {
			article, articledetial := medicine.ParseMarkdown(path, info)
			database.UpdateSummary(article)
			database.UpdateDetail(articledetial)
		}
		return nil
	},
	)
	if err != nil {
		panic(err)
	}

}

func Init() {
	type MysqlConfig struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Dbname   string `json:"dbname"`
	}
	type Config struct {
		Mysql MysqlConfig `json:"mysql"`
	}
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}
	database.InitDatabase(config.Mysql.Username, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.Dbname)
}