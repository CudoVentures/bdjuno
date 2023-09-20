#!/bin/bash -e

pkill cudos-noded || true
pkill bdjuno || true

echo "Removing NODE & BDJuno installs"
rm -rf $CUDOS_HOME
rm -rf $CUDOS_INSTALL_PATH
rm -rf $BDJUNO_INSTALL_PATH

docker rm -f $DB_NAME || true