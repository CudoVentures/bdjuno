#!/bin/bash

SETUP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "$SETUP_DIR/set_up"
chmod +x ./start_node.sh
echo "Starting the node..."
source ./start_node.sh

cd "$SETUP_DIR/set_up"
chmod +x ./start_db.sh
echo "Starting Database..."
source ./start_db.sh

cd "$SETUP_DIR/set_up"
chmod +x ./start_bdjuno.sh
echo "Starting BDJuno..."
source ./start_bdjuno.sh 

echo "Start testing..."
sleep 2
cd "$SETUP_DIR/tests"
go test -v -p 1 ./...

cd "$SETUP_DIR/set_up"
chmod +x ./clean_up.sh
echo "Clean Up..."
source ./clean_up.sh