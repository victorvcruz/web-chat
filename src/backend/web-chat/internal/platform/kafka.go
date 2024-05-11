package platform

import (
	"github.com/segmentio/kafka-go"
	"web-chat/internal/config"
)

func NewKafkaConnect(cfg config.Kafka) (*kafka.Conn, error) {
	conn, err := kafka.Dial("tcp", cfg.Broker)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
