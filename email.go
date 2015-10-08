// Copyright 2012 Santiago Corredoira
// Distributed under a BSD-like license.
package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Attachment struct {
	Filename string
	Data     []byte
	Inline   bool
	Headers  map[string]string
}

type Message struct {
	From            string
	To              []string
	Cc              []string
	Bcc             []string
	Subject         string
	Body            string
	BodyContentType string
	Attachments     map[string]*Attachment
}

func (m *Message) attach(reader io.Reader, filename string, inline bool, headers map[string]string) error {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	m.Attachments[filename] = &Attachment{
		Filename: filename,
		Data:     data,
		Inline:   inline,
		Headers:  headers,
	}

	return nil
}

func (m *Message) Attach(file string) error {
	fin, err := os.Open(file)

	if err != nil {
		return err
	}

	defer fin.Close()

	_, filename := filepath.Split(file)
	return m.attach(fin, filename, false, map[string]string{})
}

func (m *Message) AttachReader(reader io.Reader, filename string, headers map[string]string) error {
	return m.attach(reader, filename, false, headers)
}

func (m *Message) Inline(file string) error {
	fin, err := os.Open(file)

	if err != nil {
		return err
	}

	defer fin.Close()

	_, filename := filepath.Split(file)
	return m.attach(fin, filename, true, map[string]string{})
}

func (m *Message) InlineReader(reader io.Reader, filename string, headers map[string]string) error {
	return m.attach(reader, filename, true, headers)
}

func newMessage(subject string, body string, bodyContentType string) *Message {
	m := &Message{Subject: subject, Body: body, BodyContentType: bodyContentType}

	m.Attachments = make(map[string]*Attachment)

	return m
}

// NewMessage returns a new Message that can compose an email with attachments
func NewMessage(subject string, body string) *Message {
	return newMessage(subject, body, "text/plain")
}

// NewMessage returns a new Message that can compose an HTML email with attachments
func NewHTMLMessage(subject string, body string) *Message {
	return newMessage(subject, body, "text/html")
}

// ToList returns all the recipients of the email
func (m *Message) Tolist() []string {
	tolist := m.To

	for _, cc := range m.Cc {
		tolist = append(tolist, cc)
	}

	for _, bcc := range m.Bcc {
		tolist = append(tolist, bcc)
	}

	return tolist
}

// Bytes returns the mail data
func (m *Message) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("From: " + m.From + "\n")

	t := time.Now()
	buf.WriteString("Date: " + t.Format(time.RFC822) + "\n")

	buf.WriteString("To: " + strings.Join(m.To, ",") + "\n")
	if len(m.Cc) > 0 {
		buf.WriteString("Cc: " + strings.Join(m.Cc, ",") + "\n")
	}

	buf.WriteString("Subject: " + m.Subject + "\n")
	buf.WriteString("MIME-Version: 1.0\n")

	boundary := "f46d043c813270fc6b04c2d223da"

	if len(m.Attachments) > 0 {
		buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\n")
		buf.WriteString("--" + boundary + "\n")
	}

	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\n\n", m.BodyContentType))
	buf.WriteString(m.Body)
	buf.WriteString("\n")

	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			buf.WriteString("\n\n--" + boundary + "\n")
			buf.Write(attachment.Bytes())
			buf.WriteString("\n--" + boundary)
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}

func (attachment Attachment) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	if attachment.Headers == nil {
		attachment.Headers = make(map[string]string)
	}

	if attachment.Inline {
		attachment.Headers["Content-Type"] = "message/rfc822"
		attachment.Headers["Content-Disposition"] = "inline; filename=\"" + attachment.Filename + "\""
	} else {
		attachment.Headers["Content-Type"] = "application/octet-stream"
		attachment.Headers["Content-Tranfer-Encoding"] = "base64"
		attachment.Headers["Content-Disposition"] = "attachment; filename=\"" + attachment.Filename + "\""
	}

	for key, val := range attachment.Headers {
		buf.WriteString(fmt.Sprintf("%v: %v\n", key, val))
	}

	buf.Write([]byte("\n"))

	if attachment.Inline {
		buf.Write(attachment.Data)
	} else {
		b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Data)))
		base64.StdEncoding.Encode(b, attachment.Data)
		buf.Write(b)
	}

	return buf.Bytes()
}

func Send(addr string, auth smtp.Auth, m *Message) error {
	return smtp.SendMail(addr, auth, m.From, m.Tolist(), m.Bytes())
}
