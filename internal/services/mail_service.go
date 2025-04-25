package services

import (
	"bytes"
	"crypto/tls"
	"digital-marketplace/internal/models"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
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
	// baseURL := os.Getenv("BASE_URL") // Больше не нужен для ссылки

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

	// Получаем путь к файлу продукта
	fileService := NewFileService()
	productFilePath, productFileName, err := fileService.GetProductFileInfo(product.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о файле продукта: %v", err)
	}

	// Читаем файл
	fileBytes, err := ioutil.ReadFile(productFilePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла продукта '%s': %v", productFilePath, err)
	}

	bodyText := fmt.Sprintf(`Уважаемый клиент!

Спасибо за покупку в нашем Digital Marketplace!

Товар: %s
Описание: %s

Ваш продукт %s прикреплен к этому письму.

С уважением,
Команда Digital Marketplace`, product.Title, product.Description, productFileName)

	// Создаем multipart сообщение
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	// Заголовки письма
	headers := make(map[string]string)
	headers["From"] = fromEmail
	headers["To"] = to
	headers["Subject"] = fmt.Sprintf("Ваш цифровой товар: %s", product.Title)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/mixed; boundary=" + writer.Boundary()

	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buf.WriteString("\r\n")

	// Текстовая часть
	part, err := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/plain; charset=UTF-8"},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания текстовой части письма: %v", err)
	}
	part.Write([]byte(bodyText))

	// Часть с файлом (вложение)
	part, err = writer.CreatePart(map[string][]string{
		"Content-Type":              {"application/octet-stream"}, // Используем общий тип, т.к. файл - zip
		"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", productFileName)},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания части вложения: %v", err)
	}

	b64Writer := base64.NewEncoder(base64.StdEncoding, part)
	_, err = b64Writer.Write(fileBytes)
	if err != nil {
		return fmt.Errorf("ошибка записи base64 данных файла: %v", err)
	}
	b64Writer.Close() // Важно закрыть кодировщик base64

	// Закрываем multipart writer
	writer.Close()

	// Отправляем сообщение
	msgBytes := buf.Bytes()

	// Проверяем, используем ли MailHog (без TLS) или реальный SMTP-сервер (с TLS)
	if strings.ToLower(smtpHost) == "mailhog" {
		// Для MailHog используем простое SMTP без TLS и аутентификации
		err = smtp.SendMail(
			smtpHost+":"+smtpPort,
			nil, // для mailhog аутентификация не нужна
			fromEmail,
			[]string{to},
			msgBytes,
		)
		if err != nil {
			return fmt.Errorf("Ошибка отправки через MailHog: %v", err)
		}
		fmt.Printf("Письмо с вложением %s отправлено на %s через MailHog\n", productFileName, to)
		return nil
	}

	// Для реальных SMTP-серверов используем TLS
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true, // В реальном приложении должно быть false и настроен CA
		ServerName:         smtpHost,
	}

	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsconfig)
	if err != nil {
		return fmt.Errorf("ошибка TLS подключения: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("ошибка создания SMTP клиента: %v", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("ошибка аутентификации: %v", err)
	}

	if err = client.Mail(fromEmail); err != nil {
		return fmt.Errorf("ошибка команды MAIL FROM: %v", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("ошибка команды RCPT TO: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("ошибка команды DATA: %v", err)
	}

	_, err = w.Write(msgBytes)
	if err != nil {
		return fmt.Errorf("ошибка записи тела письма: %v", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("ошибка закрытия DATA writer: %v", err)
	}

	fmt.Printf("Письмо с вложением %s отправлено на %s через %s\n", productFileName, to, smtpHost)
	return client.Quit()
}
