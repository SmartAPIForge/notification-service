package service

import (
	"notification-service/internal/domain"
)

type ISenderAdapter interface {
	Send(message *domain.Message)
}
