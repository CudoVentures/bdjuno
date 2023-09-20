#!/bin/bash -e

TEST_BRANCH="cudos-dev-cosmos-v0.47.3"
export BDJUNO_INSTALL_PATH="/tmp/cudos-test-bdjuno"
export BDJUNO_CONFIG_PATH="/sample_configs/integration-tests-config/bdjuno"
HOME="$BDJUNO_INSTALL_PATH$BDJUNO_CONFIG_PATH"

# cleanup from previous runs
pkill bdjuno || true
rm -rf $BDJUNO_INSTALL_PATH

# clone cudos-node repo and install binary
git clone -b $TEST_BRANCH https://github.com/CudoVentures/cudos-bdjuno.git $BDJUNO_INSTALL_PATH
cd $BDJUNO_INSTALL_PATH
make install

bdjuno database migrate --home $HOME
bdjuno start --home $HOME &> /dev/null &
