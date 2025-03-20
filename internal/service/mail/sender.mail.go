package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"notification-service/internal/domain"
	"notification-service/internal/redis"
	"notification-service/internal/s3"
)

type SenderMail struct {
	redisClient *redis.RedisClient
	s3Client    *s3.S3Client
}

type PayloadMail struct {
	To          string
	ProjectName string
	Data        string
}

func (s *SenderMail) SendNewZipUpdate(payload *PayloadMail) {
	e := email.NewEmail()

	e.From = fmt.Sprintf("SmartAPIForge")
	e.To = []string{payload.To}
	e.Subject = "API Generated!"
	e.Text = []byte("Download .zip file:")

	file, err := s.s3Client.LoadFile(payload.Data)
	if err != nil {
		return
	}

	_, err = e.Attach(file, fmt.Sprintf("%s.zip", payload.ProjectName), "application/octet-stream")
	if err != nil {
		return
	}

	s.send(e)
}

func (s *SenderMail) SendDeployUpdate(payload *PayloadMail) {
	e := email.NewEmail()

	e.From = fmt.Sprintf("SmartAPIForge")
	e.To = []string{payload.To}
	e.Subject = "API Deployed!"
	e.Text = []byte("Your API accessible here: " + payload.Data)

	s.send(e)
}

func (s *SenderMail) send(e *email.Email) {
	fmt.Println(e)
	//addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	//auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)
	//defer e.Send(addr, auth)
}

func (s *SenderMail) FetchPayloadMail(message *domain.Message) *PayloadMail {

	email := "tyjresd@gmail.com" // TODO: from redis by message.Data["owner"].(string)

	return &PayloadMail{
		To:          email,
		ProjectName: message.Data["name"].(string),
		Data:        message.Data["url"].(string),
	}
}
