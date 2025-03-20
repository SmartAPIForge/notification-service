package mail

import (
	"notification-service/internal/domain"
	"notification-service/internal/redis"
	"notification-service/internal/s3"
)

type SenderMailAdapter struct {
	sender SenderMail
}

func NewSenderMailAdapter(
	redisClient *redis.RedisClient,
	s3Client *s3.S3Client,
) SenderMailAdapter {
	return SenderMailAdapter{
		sender: SenderMail{
			redisClient: redisClient,
			s3Client:    s3Client,
		},
	}
}

func (ms SenderMailAdapter) Send(message *domain.Message) {
	payload := ms.sender.FetchPayloadMail(message)
	switch message.Type {
	case "NewZip":
		ms.sender.SendNewZipUpdate(payload)
		break
	case "DeployPayload":
		ms.sender.SendDeployUpdate(payload)
		break
	default:
		break
	}
}
