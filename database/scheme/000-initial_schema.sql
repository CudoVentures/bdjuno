CREATE TABLE migrations 
(
    id SERIAL PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    created_at BIGINT NOT NULL
);

/* 00-cosmos.sql */

CREATE TABLE validator
(
    consensus_address TEXT NOT NULL PRIMARY KEY, /* Validator consensus address */
    consensus_pubkey  TEXT NOT NULL UNIQUE /* Validator consensus public key */
);

CREATE TABLE pre_commit
(
    validator_address TEXT                        NOT NULL REFERENCES validator (consensus_address),
    height            BIGINT                      NOT NULL,
    timestamp         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    voting_power      BIGINT                      NOT NULL,
    proposer_priority BIGINT                      NOT NULL,
    UNIQUE (validator_address, timestamp)
);

CREATE INDEX pre_commit_validator_address_index ON pre_commit (validator_address);
CREATE INDEX pre_commit_height_index ON pre_commit (height);

CREATE TABLE block
(
    height           BIGINT UNIQUE PRIMARY KEY,
    hash             TEXT                        NOT NULL UNIQUE,
    num_txs          INTEGER DEFAULT 0,
    total_gas        BIGINT  DEFAULT 0,
    proposer_address TEXT REFERENCES validator (consensus_address),
    timestamp        TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX block_height_index ON block (height);
CREATE INDEX block_hash_index ON block (hash);
CREATE INDEX block_proposer_address_index ON block (proposer_address);
ALTER TABLE block
    SET (
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_analyze_scale_factor = 0,
        autovacuum_vacuum_threshold = 10000,
        autovacuum_analyze_threshold = 10000
        );

CREATE TABLE transaction
(
    hash         TEXT    NOT NULL UNIQUE PRIMARY KEY,
    height       BIGINT  NOT NULL REFERENCES block (height),
    success      BOOLEAN NOT NULL,

    /* Body */
    messages     JSONB   NOT NULL DEFAULT '[]'::JSONB,
    memo         TEXT,
    signatures   TEXT[]  NOT NULL,

    /* AuthInfo */
    signer_infos JSONB   NOT NULL DEFAULT '[]'::JSONB,
    fee          JSONB   NOT NULL DEFAULT '{}'::JSONB,

    /* Tx response */
    gas_wanted   BIGINT           DEFAULT 0,
    gas_used     BIGINT           DEFAULT 0,
    raw_log      TEXT,
    logs         JSON
);
CREATE INDEX transaction_hash_index ON transaction (hash);
CREATE INDEX transaction_height_index ON transaction (height);

ALTER TABLE transaction
    SET (
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_analyze_scale_factor = 0,
        autovacuum_vacuum_threshold = 10000,
        autovacuum_analyze_threshold = 10000
        );
       

CREATE TABLE message
(
    transaction_hash            TEXT   NOT NULL REFERENCES transaction (hash),
    index                       BIGINT NOT NULL,
    type                        TEXT   NOT NULL,
    value                       JSONB  NOT NULL,
    involved_accounts_addresses TEXT[] NULL
);
CREATE INDEX message_transaction_hash_index ON message (transaction_hash);



/**
 * This function is used to find all the utils that involve any of the given addresses and have
 * type that is one of the specified types.
 */
CREATE FUNCTION messages_by_address(
    addresses TEXT[],
    types TEXT[],
    "limit" BIGINT = 100,
    "offset" BIGINT = 0)
    RETURNS SETOF message AS
$$
SELECT message.transaction_hash, message.index, message.type, message.value, message.involved_accounts_addresses
FROM message
         JOIN transaction t on message.transaction_hash = t.hash
WHERE (cardinality(types) = 0 OR type = ANY (types))
  AND addresses && involved_accounts_addresses
ORDER BY height DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;

CREATE TABLE pruning
(
    last_pruned_height BIGINT NOT NULL
);


/* 01-auth.sql */

CREATE TABLE account
(
    address TEXT NOT NULL PRIMARY KEY
);

/* ---- Moved from bank.sql for vesting account usage ---- */
CREATE TYPE COIN AS
(
    denom  TEXT,
    amount TEXT
);

/* ---- AUTH/ VESTING ACCOUNT ---- */
CREATE TABLE vesting_account
(
    id                  SERIAL                          PRIMARY KEY NOT NULL,
    type                TEXT                            NOT NULL,
    address             TEXT                            NOT NULL REFERENCES account (address),
    original_vesting    COIN[]                          NOT NULL DEFAULT '{}',
    end_time            TIMESTAMP WITHOUT TIME ZONE     NOT NULL,
    start_time          TIMESTAMP WITHOUT TIME ZONE
);
/* ---- start_time can be empty on DelayedVestingAccount ---- */

CREATE UNIQUE INDEX vesting_account_address_idx ON vesting_account (address);


CREATE TABLE vesting_period
(
    vesting_account_id  BIGINT  NOT NULL REFERENCES vesting_account (id),
    period_order        BIGINT  NOT NULL,
    length              BIGINT  NOT NULL,
    amount              COIN[]  NOT NULL DEFAULT '{}'
);

/* 02-bank.sql */
/* ---- SUPPLY ---- */

CREATE TABLE supply
(
    one_row_id BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    coins      COIN[]  NOT NULL,
    height     BIGINT  NOT NULL,
    CHECK (one_row_id)
);

CREATE INDEX supply_height_index ON supply (height);

/* 03-staking.sql */

/* ---- PARAMS ---- */

CREATE TABLE staking_params
(
    one_row_id BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    params     JSONB   NOT NULL,
    height     BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX staking_params_height_index ON staking_params (height);

/* ---- POOL ---- */

CREATE TABLE staking_pool
(
    one_row_id        BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    bonded_tokens     DECIMAL  NOT NULL,
    not_bonded_tokens DECIMAL  NOT NULL,
    height            DECIMAL  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX staking_pool_height_index ON staking_pool (height);

/* ---- VALIDATORS INFO ---- */

CREATE TABLE validator_info
(
    consensus_address     TEXT   NOT NULL UNIQUE PRIMARY KEY REFERENCES validator (consensus_address),
    operator_address      TEXT   NOT NULL UNIQUE,
    self_delegate_address TEXT REFERENCES account (address),
    max_change_rate       TEXT   NOT NULL,
    max_rate              TEXT   NOT NULL,
    height                BIGINT NOT NULL
);
CREATE INDEX validator_info_operator_address_index ON validator_info (operator_address);
CREATE INDEX validator_info_self_delegate_address_index ON validator_info (self_delegate_address);

CREATE TABLE validator_description
(
    validator_address TEXT   NOT NULL REFERENCES validator (consensus_address) PRIMARY KEY,
    moniker           TEXT,
    identity          TEXT,
    avatar_url        TEXT,
    website           TEXT,
    security_contact  TEXT,
    details           TEXT,
    height            BIGINT NOT NULL
);
CREATE INDEX validator_description_height_index ON validator_description (height);

CREATE TABLE validator_commission
(
    validator_address   TEXT    NOT NULL REFERENCES validator (consensus_address) PRIMARY KEY,
    commission          DECIMAL NOT NULL,
    min_self_delegation DECIMAL NOT NULL,
    height              BIGINT  NOT NULL
);
CREATE INDEX validator_commission_height_index ON validator_commission (height);

CREATE TABLE validator_voting_power
(
    validator_address TEXT   NOT NULL REFERENCES validator (consensus_address) PRIMARY KEY,
    voting_power      BIGINT NOT NULL,
    height            BIGINT NOT NULL REFERENCES block (height)
);
CREATE INDEX validator_voting_power_height_index ON validator_voting_power (height);

CREATE TABLE validator_status
(
    validator_address TEXT    NOT NULL REFERENCES validator (consensus_address) PRIMARY KEY,
    status            INT     NOT NULL,
    jailed            BOOLEAN NOT NULL,
    tombstoned        BOOLEAN NOT NULL DEFAULT FALSE,
    height            BIGINT  NOT NULL
);
CREATE INDEX validator_status_height_index ON validator_status (height);

/* ---- DOUBLE SIGN EVIDENCE ---- */

/*
 * This holds the votes that is the evidence of a double sign.
 * It should be updated on a BLOCK basis when a double sign occurs.
 */
CREATE TABLE double_sign_vote
(
    id                SERIAL PRIMARY KEY,
    type              SMALLINT NOT NULL,
    height            BIGINT   NOT NULL,
    round             INT      NOT NULL,
    block_id          TEXT     NOT NULL,
    validator_address TEXT     NOT NULL REFERENCES validator (consensus_address),
    validator_index   INT      NOT NULL,
    signature         TEXT     NOT NULL,
    UNIQUE (block_id, validator_address)
);
CREATE INDEX double_sign_vote_validator_address_index ON double_sign_vote (validator_address);
CREATE INDEX double_sign_vote_height_index ON double_sign_vote (height);

/*
 * This holds the double sign evidences.
 * It should be updated on a on BLOCK basis.
 */
CREATE TABLE double_sign_evidence
(
    height    BIGINT NOT NULL,
    vote_a_id BIGINT NOT NULL REFERENCES double_sign_vote (id),
    vote_b_id BIGINT NOT NULL REFERENCES double_sign_vote (id)
);
CREATE INDEX double_sign_evidence_height_index ON double_sign_evidence (height);



/* 04-consensus.sql */
CREATE TABLE genesis
(
    one_row_id     BOOL      NOT NULL DEFAULT TRUE PRIMARY KEY,
    chain_id       TEXT      NOT NULL,
    time           TIMESTAMP NOT NULL,
    initial_height BIGINT    NOT NULL,
    CHECK (one_row_id)
);

CREATE TABLE average_block_time_per_minute
(
    one_row_id   BOOL    NOT NULL DEFAULT TRUE PRIMARY KEY,
    average_time DECIMAL NOT NULL,
    height       BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX average_block_time_per_minute_height_index ON average_block_time_per_minute (height);

CREATE TABLE average_block_time_per_hour
(
    one_row_id   BOOL    NOT NULL DEFAULT TRUE PRIMARY KEY,
    average_time DECIMAL NOT NULL,
    height       BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX average_block_time_per_hour_height_index ON average_block_time_per_hour (height);

CREATE TABLE average_block_time_per_day
(
    one_row_id   BOOL    NOT NULL DEFAULT TRUE PRIMARY KEY,
    average_time DECIMAL NOT NULL,
    height       BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX average_block_time_per_day_height_index ON average_block_time_per_day (height);

CREATE TABLE average_block_time_from_genesis
(
    one_row_id   BOOL    NOT NULL DEFAULT TRUE PRIMARY KEY,
    average_time DECIMAL NOT NULL,
    height       BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX average_block_time_from_genesis_height_index ON average_block_time_from_genesis (height);


/* 05-mint */
/* ---- PARAMS ---- */

CREATE TABLE mint_params
(
    one_row_id BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    params     JSONB   NOT NULL,
    height     BIGINT  NOT NULL,
    CHECK (one_row_id)
);

/* ---- INFLATION ---- */

CREATE TABLE inflation
(
    one_row_id bool PRIMARY KEY DEFAULT TRUE,
    value      DECIMAL NOT NULL,
    height     BIGINT  NOT NULL,
    CONSTRAINT one_row_uni CHECK (one_row_id)
);
CREATE INDEX inflation_height_index ON inflation (height);


/* 06-distrubition.sql */
CREATE TYPE DEC_COIN AS
(
    denom  TEXT,
    amount TEXT
);

/* ---- PARAMS ---- */

CREATE TABLE distribution_params
(
    one_row_id BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    params     JSONB   NOT NULL,
    height     BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX distribution_params_height_index ON distribution_params (height);


/* ---- COMMUNITY POOL ---- */

CREATE TABLE community_pool
(
    one_row_id bool PRIMARY KEY DEFAULT TRUE,
    coins      DEC_COIN[] NOT NULL,
    height     BIGINT     NOT NULL,
    CONSTRAINT one_row_uni CHECK (one_row_id)
);
CREATE INDEX community_pool_height_index ON community_pool (height);

/* 07-pricefeed.sql  */
CREATE TABLE token
(
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE token_unit
(
    token_name TEXT NOT NULL REFERENCES token (name),
    denom      TEXT NOT NULL UNIQUE,
    exponent   INT  NOT NULL,
    aliases    TEXT[],
    price_id   TEXT
);


/* ---- TOKEN PRICES ---- */

CREATE TABLE token_price
(
    /* Needed for the below token_price function to work properly */
    id         SERIAL                      NOT NULL PRIMARY KEY,

    unit_name  TEXT                        NOT NULL REFERENCES token_unit (denom) UNIQUE,
    price      DECIMAL                     NOT NULL,
    market_cap BIGINT                      NOT NULL,
    timestamp  TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX token_price_timestamp_index ON token_price (timestamp);

CREATE TABLE token_price_history
(
    id         SERIAL                      NOT NULL PRIMARY KEY,
    unit_name  TEXT                        NOT NULL REFERENCES token_unit (denom),
    price      DECIMAL                     NOT NULL,
    market_cap BIGINT                      NOT NULL,
    timestamp  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT unique_price_for_timestamp UNIQUE (unit_name, timestamp)
);
CREATE INDEX token_price_history_timestamp_index ON token_price_history (timestamp);

/* 08-gov */
CREATE TABLE gov_params
(
    one_row_id     BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    deposit_params JSONB   NOT NULL,
    voting_params  JSONB   NOT NULL,
    tally_params   JSONB   NOT NULL,
    height         BIGINT  NOT NULL,
    CHECK (one_row_id)
);

CREATE TABLE proposal
(
    id                INTEGER   NOT NULL PRIMARY KEY,
    title             TEXT      NOT NULL,
    description       TEXT      NOT NULL,
    content           JSONB     NOT NULL,
    proposal_route    TEXT      NOT NULL,
    proposal_type     TEXT      NOT NULL,
    submit_time       TIMESTAMP NOT NULL,
    deposit_end_time  TIMESTAMP,
    voting_start_time TIMESTAMP,
    voting_end_time   TIMESTAMP,
    proposer_address  TEXT      NOT NULL REFERENCES account (address),
    status            TEXT
);
CREATE INDEX proposal_proposer_address_index ON proposal (proposer_address);

CREATE TABLE proposal_deposit
(
    proposal_id       INTEGER REFERENCES proposal (id) NOT NULL,
    depositor_address TEXT REFERENCES account (address),
    amount            COIN[],
    height            BIGINT,
    CONSTRAINT unique_deposit UNIQUE (proposal_id, depositor_address)
);
CREATE INDEX proposal_deposit_proposal_id_index ON proposal_deposit (proposal_id);
CREATE INDEX proposal_deposit_depositor_address_index ON proposal_deposit (depositor_address);
CREATE INDEX proposal_deposit_depositor_height_index ON proposal_deposit (height);

CREATE TABLE proposal_vote
(
    proposal_id   INTEGER NOT NULL REFERENCES proposal (id),
    voter_address TEXT    NOT NULL REFERENCES account (address),
    option        TEXT    NOT NULL,
    height        BIGINT  NOT NULL,
    CONSTRAINT unique_vote UNIQUE (proposal_id, voter_address)
);
CREATE INDEX proposal_vote_proposal_id_index ON proposal_vote (proposal_id);
CREATE INDEX proposal_vote_voter_address_index ON proposal_vote (voter_address);
CREATE INDEX proposal_vote_height_index ON proposal_vote (height);

CREATE TABLE proposal_tally_result
(
    proposal_id  INTEGER REFERENCES proposal (id) PRIMARY KEY,
    yes          TEXT NOT NULL,
    abstain      TEXT NOT NULL,
    no           TEXT NOT NULL,
    no_with_veto TEXT NOT NULL,
    height       BIGINT NOT NULL,
    CONSTRAINT unique_tally_result UNIQUE (proposal_id)
);
CREATE INDEX proposal_tally_result_proposal_id_index ON proposal_tally_result (proposal_id);
CREATE INDEX proposal_tally_result_height_index ON proposal_tally_result (height);

CREATE TABLE proposal_staking_pool_snapshot
(
    proposal_id       INTEGER REFERENCES proposal (id) PRIMARY KEY,
    bonded_tokens     DECIMAL NOT NULL,
    not_bonded_tokens DECIMAL NOT NULL,
    height            BIGINT NOT NULL,
    CONSTRAINT unique_staking_pool_snapshot UNIQUE (proposal_id)
);
CREATE INDEX proposal_staking_pool_snapshot_proposal_id_index ON proposal_staking_pool_snapshot (proposal_id);

CREATE TABLE proposal_validator_status_snapshot
(
    id                SERIAL PRIMARY KEY NOT NULL,
    proposal_id       INTEGER REFERENCES proposal (id),
    validator_address TEXT               NOT NULL REFERENCES validator (consensus_address),
    voting_power      BIGINT             NOT NULL,
    status            INT                NOT NULL,
    jailed            BOOLEAN            NOT NULL,
    height            BIGINT             NOT NULL,
    CONSTRAINT unique_validator_status_snapshot UNIQUE (proposal_id, validator_address)
);
CREATE INDEX proposal_validator_status_snapshot_proposal_id_index ON proposal_validator_status_snapshot (proposal_id);
CREATE INDEX proposal_validator_status_snapshot_validator_address_index ON proposal_validator_status_snapshot (validator_address);


/* 09-modules.sql */
CREATE TABLE modules
(
    module_name TEXT NOT NULL UNIQUE PRIMARY KEY
);

/* 10-slashing.sql */
CREATE TABLE validator_signing_info
(
    validator_address     TEXT                        NOT NULL PRIMARY KEY,
    start_height          BIGINT                      NOT NULL,
    index_offset          BIGINT                      NOT NULL,
    jailed_until          TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    tombstoned            BOOLEAN                     NOT NULL,
    missed_blocks_counter BIGINT                      NOT NULL,
    height                BIGINT                      NOT NULL
);
CREATE INDEX validator_signing_info_height_index ON validator_signing_info (height);

CREATE TABLE slashing_params
(
    one_row_id BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    params     JSONB   NOT NULL,
    height     BIGINT  NOT NULL,
    CHECK (one_row_id)
);
CREATE INDEX slashing_params_height_index ON slashing_params (height);


/* 11-history.sql */
CREATE TABLE account_balance_history
(
    address      TEXT       NOT NULL REFERENCES account (address),
    balance      COIN[]     NOT NULL DEFAULT '{}',
    delegated    COIN[]     NOT NULL DEFAULT '{}',
    unbonding    COIN[]     NOT NULL DEFAULT '{}',
    redelegating COIN[]     NOT NULL DEFAULT '{}',
    commission   DEC_COIN[] NOT NULL DEFAULT '{}',
    reward       DEC_COIN[] NOT NULL DEFAULT '{}',
    timestamp    TIMESTAMP WITHOUT TIME ZONE,
    CONSTRAINT unique_balance_for_height UNIQUE (address, timestamp)
);
CREATE INDEX account_balance_history_timestamp_index ON account_balance_history (timestamp);

/**
  * This function is used to have a Hasura compute field (https://hasura.io/docs/1.0/graphql/core/schema/computed-fields.html)
  * inside the account_balance table, so that it's easy to determine the token price that is associated with that balance.
 */
CREATE FUNCTION account_balance_history_tokens_prices(balance_row account_balance_history) RETURNS SETOF token_price_history AS
$$
SELECT id, unit_name, price, market_cap, timestamp
FROM (
         SELECT DISTINCT ON (unit_name) unit_name, id, price, market_cap, timestamp
         FROM (
                  SELECT *
                  FROM token_price_history
                  WHERE timestamp <= balance_row.timestamp
                  ORDER BY timestamp DESC
              ) AS prices
     ) as prices
$$ LANGUAGE sql STABLE;

/* 11-feegrant.sql */

CREATE TABLE fee_grant_allowance
(
    id                 SERIAL      NOT NULL PRIMARY KEY,
    grantee_address    TEXT        NOT NULL REFERENCES account (address),
    granter_address    TEXT        NOT NULL REFERENCES account (address),
    allowance          JSONB       NOT NULL DEFAULT '{}'::JSONB,
    height             BIGINT      NOT NULL,
    CONSTRAINT unique_fee_grant_allowance UNIQUE(grantee_address, granter_address) 
);
CREATE INDEX fee_grant_allowance_height_index ON fee_grant_allowance (height);

/* 13-cosmwasm */

CREATE TABLE cosmwasm_store
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    index BIGINT NOT NULL,
    sender TEXT NOT NULL,
    instantiate_permission JSONB DEFAULT '{}'::JSONB,
    result_code_id TEXT,
    success BOOLEAN NOT NULL,
    PRIMARY KEY(transaction_hash, index)
);

CREATE INDEX cosmwasm_store_sender_index ON cosmwasm_store (sender);
CREATE INDEX cosmwasm_store_result_code_id_index ON cosmwasm_store (result_code_id);

CREATE TABLE cosmwasm_instantiate
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    index BIGINT NOT NULL,
    admin TEXT,
    funds JSONB DEFAULT '[]'::JSONB,
    label TEXT NOT NULL,
    sender TEXT NOT NULL,
    code_id TEXT NOT NULL,
    result_contract_address TEXT,
    success BOOLEAN NOT NULL,
    PRIMARY KEY(transaction_hash, index)
);

CREATE INDEX cosmwasm_instantiate_label_index ON cosmwasm_instantiate (label);
CREATE INDEX cosmwasm_instantiate_sender_index ON cosmwasm_instantiate (sender);
CREATE INDEX cosmwasm_instantiate_code_id_index ON cosmwasm_instantiate (code_id);
CREATE INDEX cosmwasm_instantiate_result_contract_address_index ON cosmwasm_instantiate (result_contract_address);

CREATE TABLE cosmwasm_execute
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    index BIGINT NOT NULL,
    method TEXT NOT NULL,
    arguments JSONB DEFAULT '{}'::JSONB,
    funds JSONB DEFAULT '[]'::JSONB,
    sender TEXT NOT NULL,
    contract TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    PRIMARY KEY(transaction_hash, index)
);

CREATE INDEX cosmwasm_execute_method_index ON cosmwasm_execute (method);
CREATE INDEX cosmwasm_execute_sender_index ON cosmwasm_execute (sender);
CREATE INDEX cosmwasm_execute_contract_index ON cosmwasm_execute (contract);

CREATE TABLE cosmwasm_migrate
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    index BIGINT NOT NULL,
    sender TEXT NOT NULL,
    contract TEXT NOT NULL,
    code_id TEXT NOT NULL,
    arguments JSONB DEFAULT '{}'::JSONB,
    success BOOLEAN NOT NULL,
    PRIMARY KEY(transaction_hash, index)
);

CREATE INDEX cosmwasm_migrate_sender_index ON cosmwasm_migrate (sender);
CREATE INDEX cosmwasm_migrate_contract_index ON cosmwasm_migrate (contract);
CREATE INDEX cosmwasm_migrate_code_id_index ON cosmwasm_migrate (code_id);

CREATE TABLE cosmwasm_update_admin
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    index BIGINT NOT NULL,
    sender TEXT NOT NULL,
    contract TEXT NOT NULL,
    new_admin TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    PRIMARY KEY(transaction_hash, index)
);

CREATE INDEX cosmwasm_update_admin_sender_index ON cosmwasm_update_admin (sender);
CREATE INDEX cosmwasm_update_admin_contract_index ON cosmwasm_update_admin (contract);
CREATE INDEX cosmwasm_update_admin_new_admin_index ON cosmwasm_update_admin (new_admin);

CREATE TABLE cosmwasm_clear_admin
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    index BIGINT NOT NULL,
    sender TEXT NOT NULL,
    contract TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    PRIMARY KEY(transaction_hash, index)
);

CREATE INDEX cosmwasm_clear_admin_sender_index ON cosmwasm_clear_admin (sender);
CREATE INDEX cosmwasm_clear_admin_contract_index ON cosmwasm_clear_admin (contract);

CREATE TABLE proposal_vote_weighted
(
    proposal_id   INTEGER NOT NULL REFERENCES proposal (id),
    voter_address TEXT    NOT NULL REFERENCES account (address),
    option        TEXT    NOT NULL,
    weight        TEXT    NOT NULL,
    height        BIGINT  NOT NULL,
    CONSTRAINT unique_vote_weighted UNIQUE (proposal_id, voter_address, option)
);

CREATE INDEX proposal_vote_weighted_proposal_id_index ON proposal_vote_weighted (proposal_id);
CREATE INDEX proposal_vote_weighted_voter_address_index ON proposal_vote_weighted (voter_address);
CREATE INDEX proposal_vote_weighted_height_index ON proposal_vote_weighted (height);

CREATE TABLE delegation
(
    validator_address TEXT               NOT NULL REFERENCES validator_info (operator_address),
    delegator_address TEXT               NOT NULL REFERENCES account (address),
    amount            COIN               NOT NULL
);

CREATE INDEX delegation_validator_address_index ON delegation (validator_address);
CREATE INDEX delegation_delegator_address_index ON delegation (delegator_address);

CREATE TABLE gravity_orchestrator
(
    address TEXT NOT NULL PRIMARY KEY
);

CREATE TABLE gravity_transaction
(
    type TEXT NOT NULL,
    attestation_id TEXT NOT NULL UNIQUE,
    orchestrator TEXT NOT NULL REFERENCES gravity_orchestrator (address),
    receiver TEXT NOT NULL,
    votes INTEGER NOT NULL,
    consensus BOOLEAN NOT NULL,
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    height BIGINT NOT NULL REFERENCES block (height),
    PRIMARY KEY(attestation_id, orchestrator)
);

CREATE INDEX gravity_transaction_receiver_index ON gravity_transaction (receiver);
CREATE INDEX gravity_transaction_hash_index ON gravity_transaction (transaction_hash);
CREATE INDEX gravity_transaction_height_index ON gravity_transaction (height);

/*
 * This function is used to find all gravity transactions associated with given receiver
 */
CREATE FUNCTION gravity_messages_by_address(
    receiver_addr TEXT,
    "limit" BIGINT = 100,
    "offset" BIGINT = 0)
    RETURNS SETOF message AS
$$
SELECT m.transaction_hash, m.index, m.type, m.value, m.involved_accounts_addresses
FROM message m
         JOIN gravity_transaction t on m.transaction_hash = t.transaction_hash
WHERE t.receiver = receiver_addr AND t.orchestrator = ANY(m.involved_accounts_addresses)
ORDER BY t.height DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;
