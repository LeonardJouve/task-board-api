package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/LeonardJouve/task-board-api/api"
	"github.com/LeonardJouve/task-board-api/auth"
	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func main() {
	env, err := os.Open(".env")
	if err != nil {
		panic(err.Error())
	}
	oldEnv := dotenv.Load(env)
	defer oldEnv.Restore()

	if err := store.New(); err != nil {
		panic(err.Error())
	}

	app := fiber.New()

	app.All("/*", router)

	err = app.Listen(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")))
	if err != nil {
		panic(err.Error())
	}
}

func router(c *fiber.Ctx) error {
	switch strings.Split(c.Path(), "/")[1] {
	case "api":
		return api.Router(c)
	case "auth":
		return auth.Router(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}
