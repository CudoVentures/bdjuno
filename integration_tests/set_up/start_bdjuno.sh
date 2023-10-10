#!/bin/bash -e

TEST_BRANCH="automated-end-to-end-testing-v0.47.3"
export BDJUNO_INSTALL_PATH="/tmp/cudos-test-bdjuno"
export BDJUNO_CONFIG_PATH="/sample_configs/integration-tests-config/bdjuno"
TESTS_PATH="$BDJUNO_INSTALL_PATH/integration_tests/tests"
HOME="$BDJUNO_INSTALL_PATH$BDJUNO_CONFIG_PATH"

# cleanup from previous runs
pkill bdjuno || true
rm -rf $BDJUNO_INSTALL_PATH

# clone bdjuno repo and install binary
git clone -b $TEST_BRANCH https://github.com/CudoVentures/cudos-bdjuno.git $BDJUNO_INSTALL_PATH
cd $BDJUNO_INSTALL_PATH
make install

bdjuno database migrate --home $HOME
bdjuno start --home $HOME &> /dev/null &

cd $TESTS_PATH
go test -v -p 1 ./...
