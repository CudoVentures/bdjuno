#!/bin/bash
SETUP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "$SETUP_DIR"
docker-compose -p integration_e2e down -v
docker-compose -p integration_e2e up --build -d --force-recreate
docker-compose -p integration_e2e logs --follow tests

while [ "$(docker ps -q -f name=tests)" ]; do
    sleep 5
done

echo "Tests finished"
docker-compose -p integration_e2e down -v
