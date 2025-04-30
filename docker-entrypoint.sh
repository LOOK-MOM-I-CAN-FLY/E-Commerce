#!/bin/sh

echo "Waiting for database..."
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
  sleep 1
done

echo "Running migrations..."
./migrate

echo "Starting application..."
exec ./app 