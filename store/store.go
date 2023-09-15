package store

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/storage/mysql"
)

func New() *mysql.Storage {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil
	}

	return mysql.New(mysql.Config{
		Host:       os.Getenv("DB_HOST"),
		Port:       port,
		Database:   os.Getenv("DB_NAME"),
		Username:   os.Getenv("DB_USERNAME"),
		Password:   os.Getenv("DB_PASSWORD"),
		Reset:      false,
		GCInterval: 10 * time.Second,
	})
}
