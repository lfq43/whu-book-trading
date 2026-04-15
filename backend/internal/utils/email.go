package utils

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"time"

	"book-trading/backend/internal/config"
)

// SendMail 使用 TLS 连接发送 HTML 邮件
func SendMail(to, subject, body string) error {
	cfg := config.AppConfig
	host := cfg.SMTPHost
	port := cfg.SMTPPort
	addr := fmt.Sprintf("%s:%s", host, port)

	from := "WHU书籍交易网站"

	// 【修复1】163 强制要求 MAIL FROM 必须是登录邮箱本身
	// 如果你的服务器允许代发，这里请改为 from
	mailFrom := cfg.SMTPUser

	// 构建邮件内容
	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += fmt.Sprintf("Message-ID: <%s@%s>\r\n", generateMessageID(), host) // 【修复3】增加 Message-ID
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	msg += "\r\n"
	msg += body

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, host)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	// 【修复2】设置连接超时
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
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

	if err = client.Mail(mailFrom); err != nil {
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

// 生成一个简单的唯一 Message-ID
func generateMessageID() string {
	return fmt.Sprintf("%d.%d", time.Now().UnixNano(), rand.Int63())
}
