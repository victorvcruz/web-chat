package chat

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"web-chat/internal/channel"
	"web-chat/internal/event"
	"web-chat/internal/user"
)

type Service interface {
	StartChannel(ctx context.Context, userId string, channelName string) (*channel.Channel, error)
	AddUserToChannel(ctx context.Context, channelId string, userId string) error
	RemoveUserToChannel(ctx context.Context, channelId string, userId string) error
	ListChannels(ctx context.Context) ([]channel.Channel, error)
	SendMessage(ctx context.Context, chat *Chat) error
	ReceiveMessages(ctx context.Context, userId, channelId string) (chan []byte, error)
	GetChannelChats(ctx context.Context, id string) ([]Chat, error)
}

type service struct {
	channel channel.Repository
	user    user.Repository
	chat    Repository
	event   event.Client
}

func (s *service) StartChannel(ctx context.Context, userId string, channelName string) (*channel.Channel, error) {

	userModel, err := s.user.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}

	channelModel := &channel.Channel{
		Name:  channelName,
		Users: []user.User{*userModel},
	}

	err = s.channel.Create(ctx, channelModel)
	if err != nil {
		return nil, err
	}

	return channelModel, nil
}

func (s *service) AddUserToChannel(ctx context.Context, channelId string, userId string) error {

	channelModel, err := s.channel.FindById(ctx, channelId)
	if err != nil {
		return nil
	}

	userModel, err := s.user.FindById(ctx, userId)
	if err != nil {
		return nil
	}

	var userAlreadyExist bool
	for i := range channelModel.Users {
		if channelModel.Users[i].ID == userModel.ID {
			userAlreadyExist = true
		}
	}

	if !userAlreadyExist {
		channelModel.Users = append(channelModel.Users, *userModel)
	}

	err = s.channel.Update(ctx, channelModel)
	if err != nil {
		return nil
	}

	return nil
}

func (s *service) RemoveUserToChannel(ctx context.Context, channelId string, userId string) error {

	channelModel, err := s.channel.FindById(ctx, channelId)
	if err != nil {
		return nil
	}

	userModel, err := s.user.FindById(ctx, userId)
	if err != nil {
		return nil
	}

	var userExist bool
	var userIndex int
	for i := range channelModel.Users {
		if channelModel.Users[i].ID == userModel.ID {
			userExist = true
			break
		}
	}

	if !userExist {
		return nil
	}

	channelModel.Users = append(channelModel.Users[:userIndex], channelModel.Users[userIndex+1:]...)

	err = s.channel.RemoverUser(ctx, userId, channelModel.ID)
	if err != nil {
		return nil
	}

	if len(channelModel.Users) == 0 {
		err = s.DeleteFlow(ctx, channelId)
		if err != nil {
			return nil
		}
		return nil
	}

	return nil
}

func (s *service) DeleteFlow(ctx context.Context, channelId string) error {

	time.Sleep(3 * time.Second)
	channelModel, err := s.channel.FindById(ctx, channelId)
	if err != nil {
		return err
	}

	if len(channelModel.Users) == 0 {
		err = s.channel.Delete(ctx, channelId)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (s *service) ListChannels(ctx context.Context) ([]channel.Channel, error) {

	channels, err := s.channel.List(ctx)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (s *service) SendMessage(ctx context.Context, chat *Chat) error {

	userModel, err := s.user.FindById(ctx, chat.SenderId)
	if err != nil {
		return err
	}
	chat.SenderName = userModel.Username

	_, err = s.channel.FindById(ctx, chat.ChannelID)
	if err != nil {
		return err
	}

	err = s.chat.Create(ctx, chat)
	if err != nil {
		return err
	}

	err = s.event.Pub(ctx, chat, chat.ChannelID)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetChannelChats(ctx context.Context, id string) ([]Chat, error) {

	chats, err := s.chat.List(ctx, id)
	if err != nil {
		return nil, err
	}

	return chats, nil
}

func (s *service) ReceiveMessages(ctx context.Context, userId, channelId string) (chan []byte, error) {

	_, err := s.user.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}

	_, err = s.channel.FindById(ctx, channelId)
	if err != nil {
		return nil, err
	}

	chanPayload, err := s.event.Sub(ctx, userId, channelId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	chanMessages := make(chan []byte)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case payload := <-chanPayload:
				var chat Chat
				err = json.Unmarshal(payload, &chat)
				if err != nil {
					log.Println(err)
					return
				}
				chanMessages <- payload
			}
		}
	}()

	return chanMessages, nil
}

func NewService(channel channel.Repository,
	user user.Repository,
	chat Repository,
	event event.Client,
) Service {
	return &service{
		channel: channel,
		user:    user,
		chat:    chat,
		event:   event,
	}
}
