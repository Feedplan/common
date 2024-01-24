package services

import (
	"fmt"

	"github.com/go-mail/mail"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewEmailService(config SMTPConfig) *EmailService {
	es := EmailService{
		dialer: mail.NewDialer(config.Host, config.Port, config.Username, config.Password),
	}
	return &es
}

type EmailService struct {
	dialer *mail.Dialer
}

type Email struct {
	From      string
	To        []string
	Subject   string
	Plaintext string
	Embeds    []string
	HTML      string
}

func (es *EmailService) Send(email Email) error {
	msg := mail.NewMessage()
	msg.SetHeader("From", email.From)
	msg.SetHeader("To", email.To...)
	msg.SetHeader("Subject", email.Subject)
	for _, v := range email.Embeds {
		msg.Embed(v)
	}
	switch {
	case email.Plaintext != "" && email.HTML != "":
		msg.SetBody("text/plain", email.Plaintext)
		msg.AddAlternative("text/html", email.HTML)
	case email.Plaintext != "":
		msg.SetBody("text/plain", email.Plaintext)
	case email.HTML != "":
		msg.SetBody("text/html", email.HTML)
	}
	err := es.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}
