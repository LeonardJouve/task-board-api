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

	// /auth
	authGroup := app.Group("/auth")
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)
	authGroup.Get("/refresh", auth.Refresh)
	authGroup.Get("/logout", auth.Logout)

	apiGroup := app.Group("/api", auth.Protect)

	// /api/boards
	boardsGroup := apiGroup.Group("/boards")
	boardsGroup.Get("/", api.GetBoards)
	boardsGroup.Get("/:board_id/invite", api.InviteBoard)
	boardsGroup.Get("/:board_id/leave", api.LeaveBoard)
	boardsGroup.Post("/", api.CreateBoard)
	boardsGroup.Put("/", api.UpdateBoard)
	boardsGroup.Delete("/:board_id", api.DeleteBoard)

	// /api/columns
	columnsGroup := apiGroup.Group("/columns")
	columnsGroup.Get("/", api.GetColumns)
	columnsGroup.Post("/", api.CreateColumn)
	columnsGroup.Put("/", api.UpdateColumn)
	columnsGroup.Patch("/:column_id", api.MoveColumn)
	columnsGroup.Delete("/:column_id", api.DeleteColumn)

	// /api/cards
	cardsGroup := apiGroup.Group("/cards")
	cardsGroup.Get("/", api.GetCards)
	cardsGroup.Get("/:card_id/tag", api.AddTag)
	cardsGroup.Post("/", api.CreateCard)
	cardsGroup.Put("/", api.UpdateCard)
	cardsGroup.Patch("/:card_id", api.MoveCard)
	cardsGroup.Delete("/:card_id", api.DeleteCard)

	// /api/tags
	tagsGroup := apiGroup.Group("/tags")
	tagsGroup.Get("/", api.GetTags)
	tagsGroup.Post("/", api.CreateTag)
	tagsGroup.Put("/", api.UpdateTag)
	tagsGroup.Delete("/:tag_id", api.DeleteTag)

	// /api/users
	usersGroup := apiGroup.Group("/users")
	usersGroup.Get("/me", api.GetMe)

	err = app.Listen(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")))
	if err != nil {
		panic(err.Error())
	}
}
