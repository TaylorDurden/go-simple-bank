#!/bin/sh

# Exit immediately if any command in the script fails (i.e., returns a non-zero status)
set -e

# Run database migration using the migrate tool
echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

# Start the application with the provided command-line arguments
echo "start the app"
exec "$@"