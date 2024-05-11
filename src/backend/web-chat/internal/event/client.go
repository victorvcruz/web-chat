package event

import (
	"context"
	"encoding/json"
	kafkago "github.com/segmentio/kafka-go"
	"log"
	"web-chat/internal/config"
)

type Client interface {
	Pub(ctx context.Context, payload interface{}, topic string) error
	Sub(ctx context.Context, group, topic string) (chan []byte, error)
}

type event struct {
	config config.Kafka
}

func (e *event) Pub(ctx context.Context, payload interface{}, topic string) error {
	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(e.config.Broker),
		Topic:                  topic,
		Transport:              kafkago.DefaultTransport,
		AllowAutoTopicCreation: true,
	}

	marshal, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = w.WriteMessages(ctx, kafkago.Message{Value: marshal})
	if err != nil {
		log.Print("failed to write messages:", err)
	}

	if err = w.Close(); err != nil {
		log.Print("failed to close writer:", err)
	}
	return nil
}

func (e *event) Sub(ctx context.Context, group, topic string) (chan []byte, error) {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     []string{e.config.Broker},
		Topic:       topic,
		GroupID:     group,
		StartOffset: kafkago.FirstOffset,
	})

	log.Printf("listener registered for topic [%s]\n", topic)
	var messages = make(chan []byte)

	go func() {
		for {
			select {
			case <-ctx.Done():
				if err := r.Close(); err != nil {
					log.Print("failed to close reader:", err)
				}
				return
			default:
				msg, err := r.FetchMessage(ctx)
				if err != nil {
					log.Println("Failed to fetch message: " + err.Error())
				}

				if msg.Value != nil {
					log.Printf("Message received: [%s]\n", topic)
					messages <- msg.Value
					if err = r.CommitMessages(ctx, msg); err != nil {
						log.Println("Failed to commit messages")
					}
				}
			}
		}
	}()

	return messages, nil
}

func NewEvent(config config.Kafka) Client {
	return &event{
		config: config,
	}
}
