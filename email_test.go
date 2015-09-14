package email

import (
	"bytes"
	"errors"
	"net/smtp"
	"testing"
)

func TestSend(t *testing.T) {
	m := NewMessage("Hi", "this is the body")
	m.From = "to@example.com"
	m.To = []string{"to@example.com"}
	m.Cc = []string{"to@example.com", "to@example.com"}
	m.Bcc = []string{"to@example.com", "to@example.com"}

	err := m.Attach("email_test.go")
	if err != nil {
		panic(err)
	}

	err = Send("smtp.gmail.com:587", smtp.PlainAuth("", "user", "passoword", "smtp.gmail.com"), m)
	if err != nil {
		panic(err)
	}
}

func TestAttachReader(t *testing.T) {
	m := NewMessage("Hi", "this is the body")
	m.From = "to@example.com"
	m.To = []string{"to@example.com"}

	attach_content := "Testing is the future"

	m.AttachReader(bytes.NewBufferString(attach_content), "Message")

	if string(m.Attachments["Message"].Data) != attach_content {
		panic(errors.New("Content doesn't match"))
	}
}
