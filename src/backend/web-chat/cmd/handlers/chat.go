package handlers

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"log"
	"web-chat/internal/chat"
)

type Chat struct {
	service chat.Service
}

func (a *Chat) StartChannel(c *fiber.Ctx) error {
	var request map[string]string
	if err := c.BodyParser(&request); err != nil {
		return err
	}

	channelModel, err := a.service.StartChannel(c.Context(), request["user_id"], request["name"])
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(channelModel)
}

func (a *Chat) AddUserToChannel(c *fiber.Ctx) error {

	id := c.Params("id")

	userId := c.Params("userId")

	if id == "" || userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "id and userId are required"})
	}

	err := a.service.AddUserToChannel(c.Context(), id, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (a *Chat) RemoveUserToChannel(c *fiber.Ctx) error {

	id := c.Params("id")

	userId := c.Params("userId")

	if id == "" || userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "id and userId are required"})
	}

	err := a.service.RemoveUserToChannel(c.Context(), id, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (a *Chat) ListChannels(c *fiber.Ctx) error {

	channels, err := a.service.ListChannels(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(channels)
}

func (a *Chat) SendMessage(c *fiber.Ctx) error {

	var request chat.Chat
	if err := c.BodyParser(&request); err != nil {
		return err
	}

	err := a.service.SendMessage(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (a *Chat) GetChannelChats(c *fiber.Ctx) error {

	id := c.Params("id")

	chats, err := a.service.GetChannelChats(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(chats)
}

func (a *Chat) WebChat(c *websocket.Conn) {

	id := c.Params("id")

	userId := c.Params("userId")

	if id == "" || userId == "" {
		c.Close()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())

	err := a.service.AddUserToChannel(ctx, id, userId)
	if err != nil {
		log.Println(err)
		return
	}

	messages, err := a.service.ReceiveMessages(ctx, userId, id)
	if err != nil {
		log.Println(err)
		c.Close()
		return
	}

	go func() {
		defer cancel()
		defer c.Close()
		defer func() {
			err = a.service.RemoveUserToChannel(ctx, id, userId)
			if err != nil {
				log.Println(err)
				return
			}
		}()

		for {
			messageType, payload, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			switch messageType {
			case websocket.CloseMessage:
				log.Println("connection closed")
				return
			case websocket.TextMessage:
				var chatModel chat.Chat
				err = json.Unmarshal(payload, &chatModel)
				if err != nil {
					log.Println(err)
					return
				}

				err = a.service.SendMessage(ctx, &chat.Chat{
					ChannelID: id,
					SenderId:  userId,
					Message:   chatModel.Message,
				})
				if err != nil {
					log.Println(err)
					return
				}
			case websocket.PingMessage:
				err = c.WriteMessage(websocket.PongMessage, nil)
				if err != nil {
					log.Println(err)
					return
				}
			case websocket.PongMessage:
				log.Println("pong message received")
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			cancel()
			c.Close()
			return
		case message := <-messages:
			err = c.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println(err)
				cancel()
				c.Close()
				return
			}
		}
	}
}

func NewChatHandler(service chat.Service) *Chat {
	return &Chat{service: service}
}
