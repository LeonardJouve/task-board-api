package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LeonardJouve/task-board-api/api"
	"github.com/LeonardJouve/task-board-api/auth"
	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/static"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis/v3"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
		env, err := os.Open(".env")
		if err != nil {
			panic(err.Error())
		}
		oldEnv := dotenv.Load(env)
		defer oldEnv.Restore()
	}

	if err := store.Init(); err != nil {
		panic(err.Error())
	}

	if err := store.Database.AutoMigrate(
		&models.User{},
		&models.Board{},
		&models.Column{},
		&models.Card{},
		&models.Tag{},
	); err != nil {
		panic(err.Error())
	}

	schema.Init()

	app := fiber.New()

	apiGroup := app.Group("/api")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://localhost:5173", // TODO
		AllowHeaders:     "Origin, Content-Type, Accept, X-CSRF-Token, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE",
		AllowCredentials: true,
	}))

	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		ContextKey:     auth.CSRF_TOKEN,
		CookieName:     auth.CSRF_TOKEN,
		CookieDomain:   os.Getenv("HOST"),
		CookiePath:     "/",
		CookieSecure:   true,
		CookieSameSite: "Lax",
		Expiration:     time.Duration((dotenv.GetInt("CSRF_TOKEN_LIFETIME_IN_MINUTE"))) * time.Minute,
		KeyGenerator:   utils.UUIDv4,
		Storage: redis.New(redis.Config{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     dotenv.GetInt("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		}),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"id":             "api.rest.error.invalid_csrf",
				"defaultMessage": "invalid csrf token",
			})
		},
	}))

	assetsPath, err := static.Assets()
	if err != nil {
		panic(err.Error())
	}

	apiGroup.Static("/assets", assetsPath)

	// /ws
	go websocket.Process()
	apiGroup.Get("/ws", auth.Protect, websocket.HandleUpgrade, websocket.HandleSocket)

	// /auth
	authGroup := apiGroup.Group("/auth")
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)
	authGroup.Get("/refresh", auth.Refresh)
	authGroup.Get("/logout", auth.Logout)
	authGroup.Get("/csrf", auth.GetCSRF)

	restGroup := apiGroup.Group("/rest", auth.Protect)

	// /api/boards
	boardsGroup := restGroup.Group("/boards")
	boardsGroup.Get("/", api.GetBoards)
	boardsGroup.Get("/:board_id", api.GetBoard)
	boardsGroup.Get("/:board_id/invite", api.InviteBoard)
	boardsGroup.Get("/:board_id/leave", api.LeaveBoard)
	boardsGroup.Post("/", api.CreateBoard)
	boardsGroup.Put("/:board_id", api.UpdateBoard)
	boardsGroup.Delete("/:board_id", api.DeleteBoard)

	// /api/columns
	columnsGroup := restGroup.Group("/columns")
	columnsGroup.Get("/", api.GetColumns)
	columnsGroup.Get("/:column_id", api.GetColumn)
	columnsGroup.Post("/", api.CreateColumn)
	columnsGroup.Put("/:column_id", api.UpdateColumn)
	columnsGroup.Patch("/:column_id/move", api.MoveColumn)
	columnsGroup.Delete("/:column_id", api.DeleteColumn)

	// /api/cards
	cardsGroup := restGroup.Group("/cards")
	cardsGroup.Get("/", api.GetCards)
	cardsGroup.Get("/:card_id", api.GetCard)
	cardsGroup.Get("/:card_id/join", api.JoinCard)
	cardsGroup.Get("/:card_id/leave", api.LeaveCard)
	cardsGroup.Get("/:card_id/tags/:tag_id", api.AddCardTag)
	cardsGroup.Delete("/:card_id/tags/:tag_id", api.RemoveCardTag)
	cardsGroup.Post("/", api.CreateCard)
	cardsGroup.Put("/:card_id", api.UpdateCard)
	cardsGroup.Patch("/:card_id/move", api.MoveCard)
	cardsGroup.Delete("/:card_id", api.DeleteCard)

	// /api/tags
	tagsGroup := restGroup.Group("/tags")
	tagsGroup.Get("/", api.GetTags)
	tagsGroup.Get("/:tag_id", api.GetTag)
	tagsGroup.Post("/", api.CreateTag)
	tagsGroup.Put("/:tag_id", api.UpdateTag)
	tagsGroup.Delete("/:tag_id", api.DeleteTag)

	// /api/users
	usersGroup := restGroup.Group("/users")
	usersGroup.Get("/me", api.GetMe)
	usersGroup.Get("/", api.GetUsers)
	usersGroup.Get("/:user_id", api.GetUser)

	err = app.Listen(fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		panic(err.Error())
	}
}
