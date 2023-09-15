package main

import (
	"leonardjouve/api"
	"leonardjouve/auth"
	"leonardjouve/dotenv"
	"leonardjouve/store"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func main() {
	env, err := os.Open(".env")
	if err != nil {
		panic(err.Error())
	}
	oldEnv := dotenv.Load(env)
	defer oldEnv.Restore()

	app := fiber.New()

	s := store.New()
	if s == nil {
		panic("Unable to connect to the database")
	}

	app.All("/*", router)

	err = app.Listen(os.Getenv("HOST") + ":" + os.Getenv("PORT"))
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
