#!/bin/bash
SETUP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "$SETUP_DIR"
docker-compose -p integration_e2e  up --build -d 
docker-compose -p integration_e2e logs --follow tests
