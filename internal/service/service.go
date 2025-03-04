package service

import (
	"mkBlog/internal/database"
	"mkBlog/utils/medicine"
	"path/filepath"
	"time"
	"os"
	"fmt"
)

var UpdatedTime = time.Now().Format("2006-01-02 15:04:05")

// 写完文章后更新
func UpdateArticle() {
	fmt.Println("开始更新")
	err := filepath.Walk("resource", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".md" {
			//if(info.ModTime().Format("2006-01-02 15:04:05") < UpdatedTime){
				//return nil
			//}
			article, articledetial := medicine.ParseMarkdown(path)
			database.UpdateSummary(article)
			database.UpdateDetail(articledetial)
			fmt.Println("更新完成")
		}
		return nil
	},
	)
	if err != nil {
		panic(err)
	}

}