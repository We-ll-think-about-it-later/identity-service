package email

import (
	"errors"
	"net/smtp"
	"regexp"
	"strconv"
)

var (
	ErrInvalidSmtpHost = errors.New("invalid SMTP host")
	ErrInvalidSmtpPort = errors.New("invalid SMTP port")
)

type EmailSender struct {
	from               Email
	password, smtpHost string
	smtpPort           int
}

func NewEmailSender(from Email, password, smtpHost string, smtpPort int) (EmailSender, error) {
	smtpHostRegex := regexp.MustCompile(`\b(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}\b`)
	if !smtpHostRegex.MatchString(smtpHost) {
		return EmailSender{}, ErrInvalidSmtpHost
	}
	if smtpPort < 1 || smtpPort > 65535 {
		return EmailSender{}, ErrInvalidSmtpPort
	}

	return EmailSender{
		from:     from,
		password: password,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
	}, nil
}

func (sender EmailSender) Send(to Email, message string) error {
	messageBytes := []byte(message)

	auth := smtp.PlainAuth("", sender.from.Value, sender.password, sender.smtpHost)

	return smtp.SendMail(
		sender.smtpHost+":"+strconv.Itoa(sender.smtpPort),
		auth,
		sender.from.Value,
		[]string{to.Value},
		messageBytes,
	)
}
