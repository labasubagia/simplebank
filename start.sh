#!/bin/sh

set -e

# Note: Enable this if you use prod .env, not in docker compose
# echo "load env variables"
# source /app/.env

echo "start the app"
exec "$@"
