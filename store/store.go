package store

import (
	"context"
	"errors"
	"fmt"
	"os"

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

func Execute(c *fiber.Ctx, err error) bool {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
		return false
	}

	return true
}

func RollbackTransactionIfNeeded(c *fiber.Ctx, tx *gorm.DB) {
	if c.Response().StatusCode() < 400 {
		return
	}

	tx.Rollback()
}

func Init() error {
	if err := connectDatabase(); err != nil {
		return err
	}

	if err := connectRedis(); err != nil {
		return err
	}

	return nil
}

func connectDatabase() error {
	var err error
	Database, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)), &gorm.Config{})

	return err
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
