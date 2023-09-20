#!/bin/bash

# # Database details
export DB_NAME="bdjuno_test_db"
export DB_USER="postgres"
export DB_HOST="localhost"
export DB_PASS="12345"
export DB_PORT=6666

# cleanup from previous runs
docker rm -f $DB_NAME || true

docker run -d \
    --name $DB_NAME \
    -e POSTGRES_USER=$DB_USER \
    -e POSTGRES_PASSWORD=$DB_PASS \
    -e POSTGRES_DB=$DB_NAME \
    -p $DB_PORT:5432 \
    postgres
