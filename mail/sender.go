package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

//go:generate go run go.uber.org/mock/mockgen -source=sender.go -destination=./mock/sender.go
type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type GmailEmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
	smtpAuthAddress   string
	smtpServerAddress string
}

func NewGmailSender(name, fromEmailAddress, fromEmailPassword string) EmailSender {
	return &GmailEmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
		smtpAuthAddress:   "smtp.gmail.com",
		smtpServerAddress: "smtp.gmail.com:587",
	}
}

func (sender *GmailEmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, sender.smtpAuthAddress)
	return e.Send(sender.smtpServerAddress, smtpAuth)
}
