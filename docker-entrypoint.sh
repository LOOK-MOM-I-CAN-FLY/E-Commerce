#!/bin/bash
set -e

echo "Ожидание доступности базы данных..."
# Ждем пока база данных будет готова
for i in {1..30}; do
    if pg_isready -h db -U postgres -d marketplace; then
        echo "База данных готова!"
        break
    fi
    echo "Ожидание базы данных... $i"
    sleep 1
    if [ $i -eq 30 ]; then
        echo "База данных не доступна после 30 попыток"
        echo "Продолжаем запуск, но могут возникнуть проблемы с подключением"
    fi
done

# Проверка директории uploads
echo "Проверка директории uploads..."
if [ ! -d "/app/uploads" ]; then
    mkdir -p /app/uploads
    chmod 777 /app/uploads
else
    chmod 777 /app/uploads
fi

# Запуск миграций
echo "Запуск миграций..."
/app/migrate || echo "Ошибка при запуске миграций, но продолжаем работу"

# Запуск основного приложения
echo "Запуск приложения..."
exec /app/app 