package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("MAIL_FROM")
	password := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, host)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func SendEmailHTML(to string, subject string, htmlBody string) error {
	from := os.Getenv("MAIL_FROM")
	password := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\n" +
		"Subject: " + subject + "\n\n" +
		htmlBody

	auth := smtp.PlainAuth("", from, password, host)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}