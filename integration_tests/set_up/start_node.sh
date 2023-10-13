#!/bin/bash -e

TEST_BRANCH="cudos-dev-cosmos-v0.47.3"
export CUDOS_HOME="/tmp/cudos-test-data"
export CUDOS_INSTALL_PATH="/tmp/cudos-test-node"
export CHAIN_ID="test_chain"
MONIKER="test-node"
ORCH_ETH_ADDRESS=0x41D0B5762341B0FCE6aDCCF69572c663481C7286
MONITORING_ENABLED="true"
ADDR_BOOK_STRICT="false"
GRAVITY_MODULE_BALANCE="10000000000000000000000000000"
CUDOS_TOKEN_CONTRACT_ADDRESS="0xE92f6A5b005B8f98F30313463Ada5cb35500a919"
NUMBER_OF_VALIDATORS="3"
NUMBER_OF_TEST_USERS="3"
NUMBER_OF_ORCHESTRATORS="3"
VALIDATOR_BALANCE="2000000000000000000000000"
ORCHESTRATOR_BALANCE="1000000000000000000000000"
FAUCET_BALANCE="20000000000000000000000000000"
KEYRING_OS_PASS="123123123"
export BOND_DENOM="acudos"

# cleanup from previous runs
rm -rf $CUDOS_HOME
pkill cudos-noded || true
rm -rf $CUDOS_INSTALL_PATH

# clone cudos-node repo and install binary
git clone -b $TEST_BRANCH https://github.com/CudoVentures/cudos-node.git $CUDOS_INSTALL_PATH
cd $CUDOS_INSTALL_PATH
make install

# initialize node data folder
cudos-noded init $MONIKER --home=$CUDOS_HOME --chain-id=$CHAIN_ID &> /dev/null

VALID_TOKEN_CONTRACT_ADDRESS="false"
if [ "$CUDOS_TOKEN_CONTRACT_ADDRESS" = "0xE92f6A5b005B8f98F30313463Ada5cb35500a919" ] || [ "$CUDOS_TOKEN_CONTRACT_ADDRESS" = "0x12d474723cb8c02bcbf46cd335a3bb4c75e9de44" ]; then
  VALID_TOKEN_CONTRACT_ADDRESS="true"
  PARAM_UNBONDING_TIME="28800s"
  PARAM_MAX_DEPOSIT_PERIOD="21600s"
  PARAM_VOTING_PERIOD="21600s"
fi
if [ "$CUDOS_TOKEN_CONTRACT_ADDRESS" = "0x817bbDbC3e8A1204f3691d14bB44992841e3dB35" ]; then
  VALID_TOKEN_CONTRACT_ADDRESS="true"
  PARAM_UNBONDING_TIME="1814400s"
  PARAM_MAX_DEPOSIT_PERIOD="1209600s"
  PARAM_VOTING_PERIOD="432000s"
fi
if [ "$VALID_TOKEN_CONTRACT_ADDRESS" = "false" ]; then
  echo "Wrong contract address"
  exit 0;
fi;

# gas price
sed -i "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"5000000000000${BOND_DENOM}\"/" "${CUDOS_HOME}/config/app.toml"

# port 1317
# enable
sed -i "/\[api\]/,/\[/ s/enable = false/enable = true/" "${CUDOS_HOME}/config/app.toml"
sed -i "s/enabled-unsafe-cors = false/enabled-unsafe-cors = true/" "${CUDOS_HOME}/config/app.toml"

# port 9090
# enable
sed -i "/\[grpc\]/,/\[/ s/enable = false/enable = true/" "${CUDOS_HOME}/config/app.toml"

# port 26657
# enable
sed -i "s/laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/0.0.0.0:26657\"/" "${CUDOS_HOME}/config/config.toml"
sed -i "s/cors_allowed_origins = \[\]/cors_allowed_origins = \[\"\*\"\]/" "${CUDOS_HOME}/config/config.toml"

# consensus params
genesisJson=$(jq ".consensus_params.evidence.max_age_num_blocks = \"531692\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# slashing params
genesisJson=$(jq ".app_state.slashing.params.signed_blocks_window = \"19200\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.slashing.params.min_signed_per_window = \"0.1\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.slashing.params.slash_fraction_downtime = \"0.0001\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# staking params
genesisJson=$(jq ".app_state.staking.params.bond_denom = \"$BOND_DENOM\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.staking.params.unbonding_time = \"$PARAM_UNBONDING_TIME\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# crisis params
genesisJson=$(jq ".app_state.crisis.constant_fee.amount = \"5000000000000000000000\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.crisis.constant_fee.denom = \"$BOND_DENOM\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# government proposal params
genesisJson=$(jq ".app_state.gov.params.min_deposit[0].amount = \"50000000000000000000000\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gov.params.min_deposit[0].denom = \"$BOND_DENOM\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gov.params.max_deposit_period = \"$PARAM_MAX_DEPOSIT_PERIOD\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gov.params.voting_period = \"$PARAM_VOTING_PERIOD\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gov.params.quorum = \"0.5\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gov.params.threshold = \"0.5\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gov.params.veto_threshold = \"0.4\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# distribution params
genesisJson=$(jq ".app_state.distribution.params.community_tax = \"0.2\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# fractions metadata
genesisJson=$(jq ".app_state.bank.denom_metadata[0].description = \"The native staking token of the Cudos Hub\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.bank.denom_metadata[0].base = \"$BOND_DENOM\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.bank.denom_metadata[0].name = \"cudos\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.bank.denom_metadata[0].symbol = \"CUDOS\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.bank.denom_metadata[0].display = \"cudos\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

genesisJson=$(jq ".app_state.bank.denom_metadata[0].denom_units = [
  {
    \"denom\": \"acudos\",
    \"exponent\": \"0\",
    \"aliases\": [ \"attocudos\" ]
  }, {
    \"denom\": \"fcudos\",
    \"exponent\": \"3\",
    \"aliases\": [ \"femtocudos\" ]
  }, {
    \"denom\": \"pcudos\",
    \"exponent\": \"6\",
    \"aliases\": [ \"picocudos\" ]
  }, {
    \"denom\": \"ncudos\",
    \"exponent\": \"9\",
    \"aliases\": [ \"nanocudos\" ]
  }, {
    \"denom\": \"ucudos\",
    \"exponent\": \"12\",
    \"aliases\": [ \"microcudos\" ]
  }, {
    \"denom\": \"mcudos\",
    \"exponent\": \"15\",
    \"aliases\": [ \"millicudos\" ]
  }, {
    \"denom\": \"cudos\",
    \"exponent\": \"18\"
  }
]" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# gravity params
gravityId=$(echo $RANDOM | sha1sum | head -c 31)
genesisJson=$(jq ".app_state.gravity.params.gravity_id = \"$gravityId\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gravity.erc20_to_denoms[0] |= .+ {
  \"erc20\": \"$CUDOS_TOKEN_CONTRACT_ADDRESS\",
  \"denom\": \"acudos\"
}" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gravity.params.minimum_transfer_to_eth = \"1\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
genesisJson=$(jq ".app_state.gravity.params.minimum_fee_transfer_to_eth = \"1200000000000000000000\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# mint params
genesisJson=$(jq ".app_state.cudoMint.minter.norm_time_passed = \"0.53172694105988\"" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

# create zero account
(echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded keys add zero-account --home=$CUDOS_HOME --keyring-backend test > "${CUDOS_HOME}/zero-account.wallet"
chmod 600 "${CUDOS_HOME}/zero-account.wallet"
ZERO_ACCOUNT_ADDRESS=$(echo $KEYRING_OS_PASS | cudos-noded keys show zero-account -a --home=$CUDOS_HOME --keyring-backend test)
cudos-noded add-genesis-account $ZERO_ACCOUNT_ADDRESS "1${BOND_DENOM}" --home=$CUDOS_HOME

# create testUser accounts
for i in $(seq 1 $NUMBER_OF_TEST_USERS); do
    (echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded keys add "testUser$i" --home=$CUDOS_HOME --keyring-backend test > "${CUDOS_HOME}/testUser$i.wallet"
    chmod 600 "${CUDOS_HOME}/testUser$i.wallet"
    userAddress=$(echo $KEYRING_OS_PASS | cudos-noded keys show testUser$i -a --home=$CUDOS_HOME --keyring-backend test)    
    cudos-noded add-genesis-account $userAddress "10000000000000000000000000${BOND_DENOM}" --home=$CUDOS_HOME
done

# create admin account
(echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded keys add test-admin --home=$CUDOS_HOME --keyring-backend test > "${CUDOS_HOME}/test-admin.wallet"
chmod 600 "${CUDOS_HOME}/test-admin.wallet"
TEST_ADMIN_ADDRESS=$(echo $KEYRING_OS_PASS | cudos-noded keys show test-admin -a --home=$CUDOS_HOME --keyring-backend test)
cudos-noded add-genesis-account $TEST_ADMIN_ADDRESS "10cudosAdmin, 10000000000000000000000000${BOND_DENOM}" --home=$CUDOS_HOME

for i in $(seq 1 $NUMBER_OF_VALIDATORS); do
    if [ "$i" = "1" ] && [ "$ROOT_VALIDATOR_MNEMONIC" != "" ]; then
        (echo $ROOT_VALIDATOR_MNEMONIC; echo $KEYRING_OS_PASS) | cudos-noded keys add "validator-$i" --recover --home=$CUDOS_HOME --keyring-backend test
    else
        (echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded keys add "validator-$i" --home=$CUDOS_HOME --keyring-backend test > "${CUDOS_HOME}/validator-$i.wallet"
        chmod 600 "${CUDOS_HOME}/validator-$i.wallet"
    fi
    validatorAddress=$(echo $KEYRING_OS_PASS | cudos-noded keys show validator-$i -a --home=$CUDOS_HOME --keyring-backend test)
    cudos-noded add-genesis-account $validatorAddress "${VALIDATOR_BALANCE}${BOND_DENOM}" --home=$CUDOS_HOME
    cat "${CUDOS_HOME}/config/genesis.json" | jq --arg validatorAddress "$validatorAddress" '.app_state.gravity.static_val_cosmos_addrs += [$validatorAddress]' > "${CUDOS_HOME}/config/tmp_genesis.json" && mv "${CUDOS_HOME}/config/tmp_genesis.json" "${CUDOS_HOME}/config/genesis.json"
done

for i in $(seq 1 $NUMBER_OF_ORCHESTRATORS); do
    (echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded keys add "orch-$i" --home=$CUDOS_HOME --keyring-backend test > "${CUDOS_HOME}/orch-$i.wallet"
    chmod 600 "${CUDOS_HOME}/orch-$i.wallet"
    orchAddress=$(echo $KEYRING_OS_PASS | cudos-noded keys show orch-$i -a --home=$CUDOS_HOME --keyring-backend test)    
    cudos-noded add-genesis-account $orchAddress "${ORCHESTRATOR_BALANCE}${BOND_DENOM}" --home=$CUDOS_HOME
    if [ "$i" = "1" ]; then
        ORCH_01_ADDRESS="$orchAddress"
    fi
done

# # add faucet account
# if [ "$FAUCET_BALANCE" != "" ] && [ "$FAUCET_BALANCE" != "0" ]; then
#     ((echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded keys add faucet --keyring-backend test) > "${CUDOS_HOME}/faucet.wallet"
#     chmod 600 "${CUDOS_HOME}/faucet.wallet"
#     FAUCET_ADDRESS=$(echo $KEYRING_OS_PASS | cudos-noded keys show faucet -a --home=$CUDOS_HOME --keyring-backend test)
#     cudos-noded add-genesis-account $FAUCET_ADDRESS "${FAUCET_BALANCE}${BOND_DENOM}" --home=$CUDOS_HOME
# fi

# Setting gravity module account and funding it as per parameter
genesisJson=$(jq ".app_state.auth.accounts += [{
  \"@type\": \"/cosmos.auth.v1beta1.ModuleAccount\",
  \"base_account\": {
    \"account_number\": \"0\",
    \"address\": \"cudos16n3lc7cywa68mg50qhp847034w88pntq8823tx\",
    \"pub_key\": null,
    \"sequence\": \"0\"
  },
  \"name\": \"gravity\",
  \"permissions\": [
    \"minter\",
    \"burner\"
  ]
}]" "${CUDOS_HOME}/config/genesis.json")
echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"

if [ "$GRAVITY_MODULE_BALANCE" != "" ] && [ "$GRAVITY_MODULE_BALANCE" != "0" ]; then
  genesisJson=$(jq ".app_state.bank.balances += [{
    \"address\": \"cudos16n3lc7cywa68mg50qhp847034w88pntq8823tx\",
    \"coins\": [
      {
        \"amount\": \"$GRAVITY_MODULE_BALANCE\",
        \"denom\": \"acudos\"
      }
    ]
  }]" "${CUDOS_HOME}/config/genesis.json")
  echo $genesisJson > "${CUDOS_HOME}/config/genesis.json"
fi

(echo $KEYRING_OS_PASS; echo $KEYRING_OS_PASS) | cudos-noded gentx validator-1 "${VALIDATOR_BALANCE}${BOND_DENOM}" ${ORCH_ETH_ADDRESS} ${ORCH_01_ADDRESS} --min-self-delegation 2000000000000000000000000 --home=$CUDOS_HOME --chain-id $CHAIN_ID --keyring-backend test

cudos-noded collect-gentxs --home=$CUDOS_HOME

cudos-noded tendermint show-node-id --home=$CUDOS_HOME > "${CUDOS_HOME}/tendermint.nodeid"

chmod 755 "${CUDOS_HOME}/config"

# start node as daemon (in background)
cudos-noded start --home=$CUDOS_HOME &> /tmp/node.log &

# wait 8 secs and verify node is producing blocks
sleep 8
if !  &> /dev/null; then
  echo "cudos-node did not start successfully"
  exit 1
else
  echo "cudos-node started successfully"
fi
