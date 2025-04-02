package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/linkedin/goavro"
	"log/slog"
	"notification-service/internal/config"
	"notification-service/internal/domain"
	"notification-service/internal/domain/enum"
	"notification-service/internal/redis"
	"notification-service/internal/service"
)

type KafkaConsumer struct {
	log            *slog.Logger
	consumer       *kafka.Consumer
	topic          string
	codec          *goavro.Codec
	senderAdapters map[enum.SenderType]service.ISenderAdapter
	redisClient    *redis.RedisClient
}

func NewKafkaConsumer(
	log *slog.Logger,
	cfg *config.Config,
	topic string,
	codec *goavro.Codec,
	senderAdapters map[enum.SenderType]service.ISenderAdapter,
	redisClient *redis.RedisClient,
) *KafkaConsumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.KafkaHost,
		"group.id":           "notification-psg",
		"enable.auto.commit": false,
	})
	if err != nil {
		panic(fmt.Sprintf("Error creating kafka consumer %v", err))
	}

	return &KafkaConsumer{
		log:            log,
		consumer:       consumer,
		topic:          topic,
		codec:          codec,
		senderAdapters: senderAdapters,
		redisClient:    redisClient,
	}
}

func (kc *KafkaConsumer) Sub() {
	err := kc.consumer.Subscribe(kc.topic, nil)
	if err != nil {
		kc.log.Error("Error subscribing to topic: ", kc.topic, err)
	}
}

// Consume > ~1 minute wait for assigning
func (kc *KafkaConsumer) Consume() {
	switch kc.topic {
	case "NewZip":
		kc.commonConsume(kc.topic)
		break
	case "DeployPayload":
		kc.commonConsume(kc.topic)
		break
	case "NewUser":
		kc.consumeNewUser(kc.topic)
		break
	}
}

func (kc *KafkaConsumer) consumeNewUser(topic string) {
	kc.log.Info(fmt.Sprintf("Started consuming %s", topic))

	for {
		msg, err := kc.consumer.ReadMessage(-1)
		if err != nil {
			kc.log.Error(fmt.Sprintf("Error reading from topic  %s", topic), err)
			continue
		}

		kc.log.Info(fmt.Sprintf("New message from topic %s", topic))

		native, _, err := kc.codec.NativeFromTextual(msg.Value)
		if err != nil {
			kc.log.Error(fmt.Sprintf("Incorrect message while handling %s", topic), string(msg.Value), err)
			kc.commitMessage(msg)
			continue
		}

		nativeMap := native.(map[string]interface{})
		username := nativeMap["username"].(string)
		email := nativeMap["email"].(string)
		kc.redisClient.SetData("users."+username, email, nil)

		kc.commitMessage(msg)
	}
}

func (kc *KafkaConsumer) commonConsume(topic string) {
	kc.log.Info(fmt.Sprintf("Started consuming %s", topic))

	for {
		msg, err := kc.consumer.ReadMessage(-1)
		if err != nil {
			kc.log.Error(fmt.Sprintf("Error reading from topic  %s", topic), err)
			continue
		}

		kc.log.Info(fmt.Sprintf("New message from topic %s", topic))

		native, _, err := kc.codec.NativeFromTextual(msg.Value)
		if err != nil {
			kc.log.Error(fmt.Sprintf("Incorrect message while handling %s", topic), string(msg.Value), err)
			kc.commitMessage(msg)
			continue
		}

		go kc.senderAdapters[enum.Mail].Send(buildNotificationMessage(native, topic))
		//go kc.senderAdapters[enum.Telegram].Send(buildNotificationMessage(native, topic))
		kc.commitMessage(msg)
	}
}

func buildNotificationMessage(native interface{}, messageType string) *domain.Message {
	return &domain.Message{
		Type: messageType,
		Data: native.(map[string]interface{}),
	}
}

func (kc *KafkaConsumer) commitMessage(msg *kafka.Message) {
	_, err := kc.consumer.CommitMessage(msg)
	if err != nil {
		kc.log.Error(
			"Failed to commit message",
			"topic", *msg.TopicPartition.Topic,
			"partition", msg.TopicPartition.Partition,
			"offset", msg.TopicPartition.Offset,
			"error", err,
		)
	}
}
