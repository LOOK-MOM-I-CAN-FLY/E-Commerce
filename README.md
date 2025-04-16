# Digital Marketplace

## Настройка OAuth и миграция базы данных

Проект включает аутентификацию через GitHub, а также обновленную модель пользователя с полем username.

### Шаги настройки

1. Скопируйте файл `.env.example` в `.env`:
   ```
   cp .env.example .env
   ```

2. Настройте OAuth параметры в файле `.env`:
   - Для GitHub:
     - Создайте OAuth приложение на [GitHub Developer Settings](https://github.com/settings/developers)
     - Укажите URL обратного вызова: `http://localhost:8080/auth/github/callback`
     - Скопируйте Client ID и Client Secret в переменные `GITHUB_CLIENT_ID` и `GITHUB_CLIENT_SECRET`

   - Установите `OAUTH_REDIRECT_BASE` равным `http://localhost:8080` (или другому базовому URL, если вы развертываете на другом хосте)

3. Запустите миграцию для добавления поля username в базу данных:
   ```
   go run cmd/migration/main.go
   ```

4. Запустите приложение:
   ```
   go run cmd/main.go
   ```

### Новые функции

- Аутентификация через GitHub
- Поле username для пользователей
- Отображение добавленных товаров на странице профиля

### Примечания

- Для существующих пользователей поле username будет пустым, пока они не обновят свой профиль.
- При аутентификации через GitHub, имя пользователя берется из профиля GitHub. 