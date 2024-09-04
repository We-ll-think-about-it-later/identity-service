package email

import (
	"errors"
	"fmt"
	// "net/smtp"
	"regexp"
	// "strconv"
)

var (
	ErrInvalidFromEmailAddress = errors.New("invalid from email address")
	ErrInvalidToEmailAddress   = errors.New("invalid to email address")
	ErrInvalidSmtpHost         = errors.New("invalid SMTP host")
	ErrInvalidSmtpPort         = errors.New("invalid SMTP port")
)

type EmailSender struct {
	from, password, smtpHost string
	smtpPort                 int
}

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// TODO fix same types in args
func NewEmailSender(from, password, smtpHost string, smtpPort int) (EmailSender, error) {
	if !IsValidEmail(from) {
		return EmailSender{}, ErrInvalidFromEmailAddress
	}
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

func (sender EmailSender) Send(to, message string) error {
	fmt.Println(message, to)
	// if !IsValidEmail(to) {
	// 	return ErrInvalidToEmailAddress
	// }
	//
	// messageBytes := []byte(message)
	//
	// auth := smtp.PlainAuth("", sender.from, sender.password, sender.smtpHost)
	//
	// err := smtp.SendMail(
	// 	sender.smtpHost+":"+strconv.Itoa(sender.smtpPort),
	// 	auth,
	// 	sender.from,
	// 	[]string{to},
	// 	messageBytes,
	// )
	return nil
}
