package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"book-trading/backend/internal/config"
)

// SendMail 使用 TLS 连接发送 HTML 邮件
func SendMail(to, subject, body string) error {
	cfg := config.AppConfig
	host := cfg.SMTPHost
	port := cfg.SMTPPort
	addr := fmt.Sprintf("%s:%s", host, port)

	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUser
	}

	// 构建邮件内容（简单 HTML）
	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	msg += "\r\n"
	msg += body

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, host)

	// 建立 TLS 连接（适用于 465 端口）
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(cfg.SMTPUser); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
