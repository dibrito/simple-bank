#!/bin/sh

set -e

# migrations will run inside the code now
# echo "run db migration"
# source /app/app.env
# /app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"