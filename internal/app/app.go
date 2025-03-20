package app

import (
	"log/slog"
	"notification-service/internal/config"
	"notification-service/internal/domain/enum"
	"notification-service/internal/kafka"
	"notification-service/internal/redis"
	"notification-service/internal/s3"
	"notification-service/internal/service"
	"notification-service/internal/service/mail"
	"runtime/debug"
	"time"
)

type App struct {
	RedisClient *redis.RedisClient
}

func NewApp(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	redisClient := redis.NewRedisClient(cfg)
	s3Client := s3.NewS3Client(cfg)

	senderAdapters := map[enum.SenderType]service.ISenderAdapter{}
	senderAdapters[enum.Mail] = mail.NewSenderMailAdapter(redisClient, s3Client)

	schemaManager := kafka.NewSchemaManager(cfg)
	for topic, codec := range schemaManager.Schemas {
		consumer := kafka.NewKafkaConsumer(log, cfg, topic, codec, senderAdapters)
		consumer.Sub()
		go func(consumer *kafka.KafkaConsumer) {
			for {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Error("Panic in consumer, restarting...",
								"panic", r,
								"stack", string(debug.Stack()))
						}
					}()
					consumer.Consume()
				}()
				time.Sleep(5 * time.Second)
				consumer.Sub()
			}
		}(consumer)
	}

	return &App{
		RedisClient: redisClient,
	}
}
