package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"path/filepath"
)

const (
	Addr     = "smtp.126.com"
	Host     = Addr + ":25"
	AuthName = "cccbackup@126.com"
	AuthPwd  = "ndasec123"
)

type Message struct {
	Host       string   // smtp.126.com
	SenderName string   // Lucas Cai
	Sender     string   // xx@qq.com
	To         []string // {"example@qq.com","example2@qq.com"}
	ToName     []string // {"leowang","lucas"}
	Subject    string
	Body       string
	Marker     string
	Files      []*File
}

type File struct {
	fileName  string
	alterName string
}

func IsPathExists(fileName *string) (bool, error) {
	_, err := os.Stat(*fileName)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (m Message) Attach(fileName *string) error {
	exists, err := IsPathExists(fileName)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("Path not exist!")
	}

	m.Files = append(m.Files, &File{fileName: *fileName, alterName: *fileName})
	return nil
}

func NewMessage() *Message {
	return &Message{
		Host:       "smtp.126.com",
		SenderName: "cnt",
		Sender:     "cccbackup@126.com",
		To:         []string{"727266990@qq.com"},
		ToName:     []string{"LeoCai"},
		Subject:    "ONLY",
		Body:       "You are my distiny",
		Marker:     "5fab5a3e4219c2e3a186fd32b610a146bf1b8609fff08cf38d0ddfb10a1a",
	}
}

// mail headers
func (m *Message) Head() []byte {
	return []byte(fmt.Sprintf("From: %s <%s>\r\nTo: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed;\r\n boundary=%s\r\n\r\n--%s",
		m.SenderName, m.Sender, m.ToName[0], m.To[0], m.Subject, m.Marker, m.Marker))
}

// body (text or HTML)
func (m *Message) Bodys() []byte {
	return []byte(fmt.Sprintf("\r\nContent-Type: text/html\r\nContent-Transfer-Encoding:8bit\r\n\r\n%s\r\n--%s",
		m.Body, m.Marker))
}

func (m *Message) Encode(fileName *string) ([]byte, error) {
	var buf bytes.Buffer
	name := filepath.Base(*fileName)
	content, err := ioutil.ReadFile(*fileName)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(content)

	lineMaxLength := 500
	nbrLines := len(encoded) / lineMaxLength

	// append lines to buffer
	for i := 0; i < nbrLines; i++ {
		buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
	}

	// append last line in buffer
	buf.WriteString(encoded[nbrLines*lineMaxLength:])
	var rst string
	rst += "\r\n"
	rst += "Content-Type: application/octet-stream;charset=utf-8; name=\"%s\"\r\n"
	rst += "Content-Transfer-Encoding: base64\r\n"
	rst += "Content-Disposition: attachment; charset=utf-8; filename=\"%s\"\r\n"
	rst += "\r\n%s\r\n"
	rst += "--%s--"

	return []byte(fmt.Sprintf(rst, name, name, buf.String(), m.Marker)), nil
}

func (m *Message) ToBytes() ([]byte, error) {
	encodeBytes := make([]byte, 0)
	encodeBytes = append(encodeBytes, m.Head()...)
	encodeBytes = append(encodeBytes, m.Bodys()...)
	for i := 0; i < len(m.Files); i++ {
		bytes, err := m.Encode(&m.Files[i].alterName)
		if err != nil {
			return nil, err
		}

		encodeBytes = append(encodeBytes, bytes...)
	}

	return encodeBytes, nil
}

func SendMail(auth *smtp.Auth, m *Message) error {
	ctx, err := m.ToBytes()
	if err != nil {
		return err
	}

	if err := smtp.SendMail(m.Host, *auth, m.Sender, m.To, ctx); err != nil {
		return err
	}

	return nil
}

func main() {
	// gen attach
	fileName := "文件"
	f, err := os.Create(fileName)
	defer os.Remove(f.Name())
	defer f.Close()
	if err != nil {
		panic(err)
	}

	f.WriteString("实践出真知")
	m := NewMessage()
	m.Attach(&fileName)
	//	m.Attach(&fileName)
	auth := smtp.PlainAuth(m.Sender, AuthName, AuthPwd, Addr)
	// send email
	count := 4
	finish := make(chan bool)
	//	mail, _ := os.Create("mail.txt")
	//	mail.WriteString(msg)
	for i := 0; i < count; i++ {
		go func() {
			defer func() { finish <- true }()
			//send the email
			if err := SendMail(&auth, m); err != nil {
				return
			}
		}()
	}

	for i := 0; i < count; i++ {
		<-finish
	}
}
