package main

import (
	"fmt"
	"os"

	"github.com/LeonardJouve/task-board-api/api"
	"github.com/LeonardJouve/task-board-api/auth"
	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	app.Use(cors.New(cors.Config{
		AllowOrigins:     fmt.Sprintf("http://%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE",
		AllowCredentials: true,
	}))

	app.All("/api/*", auth.Protect, api.Router)
	app.All("/auth/*", auth.Router)

	err = app.Listen(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")))
	if err != nil {
		panic(err.Error())
	}
}
