package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"log"
	"web-chat/cmd/handlers"
	"web-chat/internal/channel"
	"web-chat/internal/chat"
	"web-chat/internal/config"
	"web-chat/internal/event"
	"web-chat/internal/platform"
	"web-chat/internal/user"
)

func init() {
	err := godotenv.Load("src/backend/web-chat/.env")
	if err != nil {
		log.Printf("[LOAD ENVIRONMENT VARIABLES FAIL]: %s\n", err.Error())
	}
}

func main() {
	cfg := config.Load()

	connect, err := platform.NewPostgresConnect(cfg.Database)
	if err != nil {
		log.Fatalf("[CONNECT DATABASE FAIL]: %s", err.Error())
	}

	err = platform.Migrate(connect, user.User{}, &channel.Channel{}, &chat.Chat{})
	if err != nil {
		log.Fatalf("[MIGRATE DATABASE FAIL]: %s", err.Error())
	}

	eventClient := event.NewEvent(cfg.Kafka)

	userRepository := user.NewRepository(connect)

	channelRepository := channel.NewRepository(connect)

	chatRepository := chat.NewRepository(connect)

	chatService := chat.NewService(channelRepository, userRepository, chatRepository, eventClient)

	userHandler := handlers.NewUserHandler(userRepository)

	chatHandler := handlers.NewChatHandler(chatService)

	// Routes
	app := fiber.New()
	app.Use(logger.New(), cors.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Post("/users", userHandler.CreateUser)
	app.Put("/users/:id", userHandler.UpdateUser)
	app.Get("/users/:id", userHandler.FindUser)
	app.Delete("/users/:id", userHandler.DeleteUser)

	app.Post("/chat", chatHandler.StartChannel)
	app.Get("/chat", chatHandler.ListChannels)
	app.Post("/chat/:id/user/:userId", chatHandler.AddUserToChannel)
	app.Delete("/chat/:id/user/:userId", chatHandler.RemoveUserToChannel)

	app.Post("/chat/message", chatHandler.SendMessage)

	app.Get("/ws/chat/:id/user/:userId", websocket.New(chatHandler.WebChat))
	app.Get("/chat/channel/:id", chatHandler.GetChannelChats)

	if err = app.Listen(":9090"); err != nil {
		log.Fatalf("[START SERVER FAIL]: %s", err.Error())
	}
}
