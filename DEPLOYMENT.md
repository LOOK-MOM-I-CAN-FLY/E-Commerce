# Руководство по деплою Digital Marketplace

В этом документе описаны шаги по настройке деплоя Digital Marketplace с использованием Docker, CI/CD на базе GitHub Actions, настройка SMTP и GitHub OAuth.

## 1. Настройка окружения

### 1.1. Файл .env

Скопируйте файл `example.env` в `.env` и отредактируйте следующие параметры:

```bash
cp example.env .env
```

Основные параметры для настройки:

#### База данных
```
DB_USER=postgres
DB_PASSWORD=strong_password_here
DB_NAME=marketplace
DB_PORT_EXTERNAL=5433
```

#### SMTP для отправки писем
```
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASS=your_smtp_password
SMTP_FROM_EMAIL=noreply@example.com
```

#### GitHub OAuth
```
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
OAUTH_REDIRECT_BASE=https://your-domain.com
```

#### Настройки CI/CD
```
REGISTRY_URL=ghcr.io
REGISTRY_USER=your-github-username
REGISTRY_PASS=your-github-token
```

### 1.2. SSL-сертификаты

Для работы по HTTPS, поместите SSL-сертификаты в директорию `nginx/ssl/`:
- `fullchain.pem` - полная цепочка сертификатов
- `privkey.pem` - приватный ключ

После этого в `nginx/marketplace.conf` раскомментируйте строку редиректа с HTTP на HTTPS:
```
return 301 https://$host$request_uri;
```

## 2. Запуск в Docker

### 2.1. Локальный запуск

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 2.2. Персистентное хранение данных

Все данные хранятся в именованных Docker-томах:
- `marketplace_pgdata` - данные базы данных PostgreSQL
- `marketplace_uploads` - загруженные файлы пользователей

Для резервного копирования базы данных используется отдельный сервис `backup`.

### 2.3. Резервное копирование

Резервные копии создаются автоматически каждые 24 часа и хранятся в директории `backups/`.

Ручное создание резервной копии:
```bash
docker-compose exec db pg_dump -U postgres -d marketplace -F c -f /tmp/backup.sql
docker cp $(docker-compose ps -q db):/tmp/backup.sql ./backups/manual_backup.sql
```

Восстановление из резервной копии:
```bash
docker cp ./backups/backup_file.sql $(docker-compose ps -q db):/tmp/backup.sql
docker-compose exec db pg_restore -U postgres -d marketplace -c /tmp/backup.sql
```

## 3. CI/CD с GitHub Actions

### 3.1. Настройка GitHub Secrets

В настройках вашего GitHub репозитория добавьте следующие секреты:

- `REGISTRY_USER` - имя пользователя GitHub
- `REGISTRY_PASS` - токен доступа GitHub с правами на загрузку пакетов
- `VPS_HOST` - IP-адрес вашего сервера
- `VPS_USER` - имя пользователя для подключения к серверу
- `VPS_SSH_KEY` - приватный SSH-ключ для доступа к серверу
- `ENV_FILE` - содержимое файла .env

### 3.2. Процесс CI/CD

Рабочий процесс CI/CD включает следующие этапы:
1. Тестирование приложения
2. Сборка Docker-образа
3. Публикация Docker-образа в GitHub Container Registry
4. Деплой на сервер через SSH
5. Создание резервной копии перед обновлением
6. Обновление приложения на сервере
7. Проверка работоспособности приложения

## 4. Настройка SMTP

В файле `.env` настройте параметры SMTP-сервера:

```
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASS=your_smtp_password
SMTP_FROM_EMAIL=noreply@example.com
```

## 5. Настройка GitHub OAuth

### 5.1. Создание OAuth приложения в GitHub

1. Перейдите на страницу [GitHub Developer Settings](https://github.com/settings/developers)
2. Создайте новое OAuth приложение
3. Укажите URL обратного вызова: `https://your-domain.com/auth/github/callback`
4. Скопируйте Client ID и Client Secret

### 5.2. Настройка приложения

Добавьте полученные данные в файл `.env`:

```
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
OAUTH_REDIRECT_BASE=https://your-domain.com
```

## 6. Мониторинг и обслуживание

### 6.1. Проверка работоспособности

Приложение предоставляет эндпоинт `/health` для проверки работоспособности.

Проверка:
```bash
curl -f http://your-domain.com/health
```

Должен вернуть `OK` со статусом 200.

### 6.2. Просмотр логов

```bash
# Логи приложения
docker-compose logs -f app

# Логи базы данных
docker-compose logs -f db

# Логи Nginx
docker-compose logs -f nginx
```

### 6.3. Поддержка безопасности

- Регулярно обновляйте Docker-образы
- Используйте сложные пароли для базы данных и SMTP
- Держите SSL-сертификаты актуальными
- Настройте файервол для ограничения доступа

## 7. Устранение неполадок

### 7.1. Приложение не запускается

Проверьте логи:
```bash
docker-compose logs app
```

### 7.2. Проблемы с базой данных

Проверьте статус подключения:
```bash
docker-compose exec app ping db
docker-compose exec db pg_isready
```

### 7.3. Проблемы с SMTP

Проверьте подключение к SMTP-серверу:
```bash
docker-compose exec app nc -zv $SMTP_HOST $SMTP_PORT
```

### 7.4. Восстановление после сбоя

```bash
# Остановить все контейнеры
docker-compose down

# Запустить заново
docker-compose up -d
``` 