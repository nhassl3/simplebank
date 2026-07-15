#!/bin/sh
set -e

echo "waiting for postgres..."

until /app/migrate -database "$CONNECTION_STRING" -path /app/migration -verbose up 2>/dev/null
do
  echo "postgres is unavailable - sleeping"
  sleep 2
done

echo "db migration completed"

echo "start the application"
exec "$@"
