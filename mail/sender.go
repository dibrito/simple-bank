package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attchedFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string,
	fromEmailAddress string,
	fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}

}

func (gs *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attchedFiles []string,
) error {

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", gs.name, gs.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attchedFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("attach file %s:%w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", gs.fromEmailAddress, gs.fromEmailPassword, smtpAuthAddress)
	e.Send(smtpServerAddress, smtpAuth)

	return e.Send(smtpServerAddress, smtpAuth)
}
