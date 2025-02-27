package database

import (
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
	"mkBlog/models"
)

var Db *gorm.DB

func InitDatabase(){
	dsn := "root:root@tcp(localhost:3306)/mkblog?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil{
		panic("failed to connect database")
	}
	Db.AutoMigrate(&models.ArticleSummary{}, &models.ArticleDetail{})
}

func UpdateSummary(article models.ArticleSummary){
	var tmp models.ArticleSummary
	result := Db.Where("title = ?", article.Title).First(&tmp)
	if result != nil {
		if result.Error == gorm.ErrRecordNotFound {
			Db.Create(&article)
		}else{
			Db.Model(&tmp).Updates(article)
		}
	}
	Db.Create(&article)
}

func UpdateDetail(articleDetail models.ArticleDetail){
	var tmp models.ArticleDetail
	result := Db.Where("title = ?", articleDetail.Title).First(&tmp)
	if result != nil {
		if result.Error == gorm.ErrRecordNotFound {
			Db.Create(&articleDetail)
		}else{
			Db.Model(&tmp).Updates(articleDetail)
		}
	}
	Db.Create(&articleDetail)
}
