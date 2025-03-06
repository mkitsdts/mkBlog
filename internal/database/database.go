package database

import (
	"errors"
	"mkBlog/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDatabase(username string, password string, host string, port string, dbname string) {
	dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil{
		panic("failed to connect database")
	}
	Db.AutoMigrate(&models.ArticleSummary{}, &models.ArticleDetail{}, &models.FriendApplyment{}, &models.Friend{})
}

func UpdateSummary(article models.ArticleSummary) error {
	var tmp models.ArticleSummary
	result := Db.Where("title = ?", article.Title).First(&tmp)
	if result != nil {
		if result.Error == gorm.ErrRecordNotFound {
			Db.Create(&article)
			return nil;
		}else{
			Db.Model(&tmp).Updates(article)
			return nil;
		}
	}
	return errors.New("unknown error")
}

func UpdateDetail(articleDetail models.ArticleDetail) error {
	var tmp models.ArticleDetail
	result := Db.Where("title = ?", articleDetail.Title).First(&tmp)
	if result != nil {
		if result.Error == gorm.ErrRecordNotFound {
			Db.Create(&articleDetail)
			return nil;
		}else{
			Db.Model(&tmp).Updates(articleDetail)
			return nil;
		}
	}
	return errors.New("unknown error")
}

func ApplyFriend(friendApplyment models.FriendApplyment) error{
	var tmp models.FriendApplyment
	result := Db.Where("name = ?", friendApplyment.Name).First(&tmp)
	if result != nil {
		if result.Error == gorm.ErrRecordNotFound {
			Db.Create(&friendApplyment)
			return nil;
		}else{
			Db.Model(&tmp).Updates(friendApplyment)
			return nil;
		}
	}
	return errors.New("unknown error")
}