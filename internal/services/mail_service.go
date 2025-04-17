package services

import (
	"crypto/tls"
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
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	baseURL := os.Getenv("BASE_URL")

	if smtpHost == "" || smtpPort == "" {
		fmt.Println("SMTP_HOST или SMTP_PORT не установлены. Пропускаем отправку email.")
		return nil
	}

	// Для Mailhog не требуются учетные данные
	if smtpHost != "mailhog" && (smtpUser == "" || smtpPass == "") {
		fmt.Println("SMTP_USER или SMTP_PASS не установлены. Пропускаем отправку email.")
		return nil
	}

	if fromEmail == "" {
		fromEmail = "marketplace@example.com"
		if smtpUser != "" {
			fromEmail = smtpUser
		}
	}

	// Если BASE_URL не установлен, используем localhost по умолчанию
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Создаем защищенный токен для скачивания вместо прямой ссылки
	fileService := NewFileService()
	downloadToken, err := fileService.GenerateDownloadToken(product.ID)
	if err != nil {
		return fmt.Errorf("ошибка создания токена скачивания: %v", err)
	}

	// Формируем безопасную ссылку на скачивание с использованием токена
	downloadURL := fileService.GenerateDownloadURL(downloadToken, baseURL)

	body := fmt.Sprintf(`Уважаемый клиент!

Спасибо за покупку в нашем Digital Marketplace!

Товар: %s
Описание: %s

Ссылка на скачивание (действительна 24 часа):
%s

С уважением,
Команда Digital Marketplace`, product.Title, product.Description, downloadURL)

	msg := "From: " + fromEmail + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Ваш цифровой товар из Digital Marketplace\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	// Проверяем, используем ли MailHog (без TLS) или реальный SMTP-сервер (с TLS)
	if strings.ToLower(smtpHost) == "mailhog" {
		// Для MailHog используем простое SMTP без TLS и аутентификации
		err = smtp.SendMail(
			smtpHost+":"+smtpPort,
			nil, // для mailhog аутентификация не нужна
			fromEmail,
			[]string{to},
			[]byte(msg),
		)
		if err != nil {
			return fmt.Errorf("Ошибка отправки через MailHog: %v", err)
		}
		return nil
	}

	// Для реальных SMTP-серверов используем TLS
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpHost,
	}

	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsconfig)
	if err != nil {
		return fmt.Errorf("Ошибка TLS подключения: %v", err)
	}

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("Ошибка создания SMTP клиента: %v", err)
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("Ошибка аутентификации: %v", err)
	}

	if err = client.Mail(fromEmail); err != nil {
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
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
