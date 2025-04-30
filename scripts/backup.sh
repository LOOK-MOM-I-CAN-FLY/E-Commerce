#!/bin/bash

# Создаем директорию для резервных копий, если она не существует
mkdir -p /backups

# Устанавливаем имя файла с датой
BACKUP_FILE="/backups/marketplace_$(date +%Y%m%d_%H%M%S).sql"

# Выполняем резервное копирование
echo "Creating backup to $BACKUP_FILE"
pg_dump -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -F c -b -v -f "$BACKUP_FILE"

# Удаляем старые резервные копии (оставляем только 7 последних)
echo "Cleaning old backups..."
ls -tp /backups/*.sql | grep -v '/$' | tail -n +8 | xargs -I {} rm -- {}

echo "Backup completed successfully!"

# Ждем 24 часа перед следующим резервным копированием
echo "Waiting for next backup cycle..."
sleep 24h

# Перезапускаем скрипт
exec "$0" 