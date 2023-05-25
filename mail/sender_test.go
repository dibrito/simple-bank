package mail

import (
	"testing"

	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(c.EmailSenderName, c.EmailSenderAddress, c.EmailSenderPassword)
	subject := "Test email"
	content := `
	<h1>Hellow sinner</h1>
	<p>This is a message from hell!</p>
	`
	to := []string{"pdibrito@gmail.com"}
	attachedFiles := []string{"../README.md"}
	err = sender.SendEmail(subject, content, to, nil, nil, attachedFiles)
	require.NoError(t, err)
}
