package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"notification-service/internal/config"
	"notification-service/internal/domain"
	"notification-service/internal/redis"
	"notification-service/internal/s3"
)

type SenderMail struct {
	cfg         *config.Config
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

	e.From = fmt.Sprintf("a89872858202@yandex.ru")
	e.To = []string{payload.To}
	e.Subject = "SmartAPIForge API Generated!"
	e.Text = []byte("")

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

	e.From = fmt.Sprintf("a89872858202@yandex.ru")
	e.To = []string{payload.To}
	e.Subject = "SmartAPIForge API Deployed!"
	e.Text = []byte(fmt.Sprintf("You can access our API directly at %s. Start integrating now!", payload.Data))

	s.send(e)
}

func (s *SenderMail) send(e *email.Email) {
	fmt.Println(e)
	addr := fmt.Sprintf("%s:%s", "smtp.yandex.ru", "587")
	auth := smtp.PlainAuth("", s.cfg.SmtpLogin, s.cfg.SmtpPassword, "smtp.yandex.ru")
	defer func(e *email.Email, addr string, a smtp.Auth) {
		err := e.Send(addr, a)
		if err != nil {
			fmt.Println("Can not send message: " + err.Error())
		}
	}(e, addr, auth)
}

func (s *SenderMail) FetchPayloadMail(message *domain.Message) *PayloadMail {

	mail, _ := s.redisClient.GetData("users." + message.Data["owner"].(string))

	return &PayloadMail{
		To:          mail,
		ProjectName: message.Data["name"].(string),
		Data:        message.Data["url"].(string),
	}
}
