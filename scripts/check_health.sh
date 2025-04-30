#!/bin/bash

# Скрипт для проверки работоспособности приложения
# Использование: ./scripts/check_health.sh [хост]

# Устанавливаем хост по умолчанию
HOST=${1:-"http://localhost"}

echo "Проверка работоспособности для $HOST..."

# Проверяем доступность эндпоинта /health
response=$(curl -s -o /dev/null -w "%{http_code}" $HOST/health)

if [ $response -eq 200 ]; then
    echo "✅ Приложение работает нормально (код ответа: $response)"
    exit 0
else
    echo "❌ Ошибка: приложение недоступно (код ответа: $response)"
    
    # Проверяем запущены ли контейнеры
    if command -v docker &> /dev/null; then
        echo "Проверка статуса контейнеров Docker..."
        docker-compose ps
        
        echo "Логи приложения за последние 50 строк:"
        docker-compose logs --tail=50 app
    fi
    
    exit 1
fi 