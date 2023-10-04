package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Database *gorm.DB
	Redis    *redis.Client
)

func BeginTransaction(c *fiber.Ctx) (*gorm.DB, bool) {
	tx := Database.Begin()
	if tx.Error != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return nil, false
	}

	return tx, true
}

func Execute(c *fiber.Ctx, tx *gorm.DB, err error) bool {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
		return false
	}

	return true
}

func New() error {
	if err := connectDatabase(); err != nil {
		return err
	}

	if err := connectRedis(); err != nil {
		return err
	}

	return nil
}

func connectDatabase() error {
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

	Database.AutoMigrate(&models.User{})
	Database.AutoMigrate(&models.Board{})
	Database.AutoMigrate(&models.Column{})
	Database.AutoMigrate(&models.Card{})
	Database.AutoMigrate(&models.Tag{})

	return nil
}

func connectRedis() error {
	ctx := context.TODO()

	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	if _, err := Redis.Ping(ctx).Result(); err != nil {
		return err
	}

	return nil
}
