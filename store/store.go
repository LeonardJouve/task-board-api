package store

import (
	"fmt"
	"os"

	"github.com/LeonardJouve/task-board-api/store/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB

func New() error {
	DB, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)), &gorm.Config{})

	Database = DB

	if err != nil {
		return err
	}

	Database.AutoMigrate(&models.Board{})
	Database.AutoMigrate(&models.Column{})
	Database.AutoMigrate(&models.Card{})
	Database.AutoMigrate(&models.Tag{})

	return nil
}
