#!/bin/sh

set -e

# Note: Enable this if you use prod .env, not in docker compose
# echo "load env variables"
# source /app/.env

echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"