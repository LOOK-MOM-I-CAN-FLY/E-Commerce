
package services

import (
    "digital-marketplace/internal/models"
    "fmt"
    "net/smtp"
    "os"
    "strings"
)

func SendProductToEmail(to string, product models.Product) error {
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")

    auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

    subject := "Ваш цифровой товар"
    body := fmt.Sprintf("Спасибо за покупку! Ваш товар: %s\nСсылка на файл: http://localhost:8080/%s", product.Title, strings.TrimPrefix(product.FilePath, "./"))

    msg := "To: " + to + "\r\n" +
        "Subject: " + subject + "\r\n\r\n" +
        body

    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{to}, []byte(msg))
    return err
}
