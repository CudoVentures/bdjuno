-- DROPS
DROP FUNCTION IF EXISTS account_balance_history_tokens_prices;
DROP TABLE IF EXISTS account_balance_history;
DROP FUNCTION IF EXISTS account_balance_tokens_prices;
DROP TABLE IF EXISTS account_balance;
DROP TABLE IF EXISTS delegation_reward;
DROP TABLE IF EXISTS delegators_to_refresh;
DROP TABLE IF EXISTS nft_mint;
DROP TABLE IF EXISTS nft_issue_denom;
DROP TABLE IF EXISTS redelegation;
DROP TABLE IF EXISTS unbonding_delegation;
DROP TABLE IF EXISTS validator_commission_amount;

-- 00-cosmos.sql
UPDATE transaction
SET logs = REPLACE(logs::text, '\u0000', '')::json
WHERE logs::text LIKE '%\u0000%';

CREATE TABLE transaction_new
(
    hash         TEXT    NOT NULL,
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
    logs         JSONB,

    /* PSQL partition */
    partition_id BIGINT  NOT NULL DEFAULT 0,

    CONSTRAINT unique_tx UNIQUE (hash, partition_id)
)PARTITION BY LIST(partition_id);

CREATE TABLE transaction_0 PARTITION OF transaction_new FOR VALUES IN (0);

-- Copy data from the old table to the new partitioned table
INSERT INTO transaction_new SELECT * FROM transaction;

-- Rename tables
ALTER TABLE transaction RENAME TO transaction_old;
ALTER TABLE transaction_new RENAME TO transaction;

-- remove all foreign keys
ALTER TABLE cosmwasm_store DROP CONSTRAINT cosmwasm_store_transaction_hash_fkey;
ALTER TABLE cosmwasm_instantiate DROP CONSTRAINT cosmwasm_instantiate_transaction_hash_fkey;
ALTER TABLE cosmwasm_execute DROP CONSTRAINT cosmwasm_execute_transaction_hash_fkey;
ALTER TABLE cosmwasm_migrate DROP CONSTRAINT cosmwasm_migrate_transaction_hash_fkey;
ALTER TABLE cosmwasm_update_admin DROP CONSTRAINT cosmwasm_update_admin_transaction_hash_fkey;
ALTER TABLE cosmwasm_clear_admin DROP CONSTRAINT cosmwasm_clear_admin_transaction_hash_fkey;
ALTER TABLE gravity_transaction DROP CONSTRAINT gravity_transaction_transaction_hash_fkey;
ALTER TABLE nft_denom DROP CONSTRAINT nft_denom_transaction_hash_fkey;
ALTER TABLE nft_nft DROP CONSTRAINT nft_nft_transaction_hash_fkey;
ALTER TABLE distinct_message DROP CONSTRAINT distinct_message_transaction_hash_fkey;
ALTER TABLE group_proposal DROP CONSTRAINT group_proposal_transaction_hash_fkey;
ALTER TABLE marketplace_collection DROP CONSTRAINT marketplace_collection_transaction_hash_fkey;
ALTER TABLE marketplace_nft DROP CONSTRAINT marketplace_nft_transaction_hash_fkey;
ALTER TABLE marketplace_nft_buy_history DROP CONSTRAINT marketplace_nft_buy_history_transaction_hash_fkey;
ALTER TABLE nft_transfer_history DROP CONSTRAINT nft_transfer_history_transaction_hash_fkey;
ALTER TABLE message DROP CONSTRAINT message_transaction_hash_fkey;

-- Drop the old table
DROP TABLE transaction_old;

-- Create indexes on the new partitioned table
CREATE INDEX transaction_hash_index ON transaction (hash);
CREATE INDEX transaction_height_index ON transaction (height);
CREATE INDEX transaction_partition_id_index ON transaction (partition_id);

-- MESSAGE
CREATE TABLE message_new
(
    transaction_hash            TEXT   NOT NULL,
    index                       BIGINT NOT NULL,
    type                        TEXT   NOT NULL,
    value                       JSONB  NOT NULL,
    involved_accounts_addresses TEXT[] NOT NULL,

    /* PSQL partition */
    partition_id                BIGINT NOT NULL DEFAULT 0,
    height                      BIGINT NOT NULL,
    FOREIGN KEY (transaction_hash, partition_id) REFERENCES transaction (hash, partition_id),
    CONSTRAINT unique_message_per_tx UNIQUE (transaction_hash, index, partition_id)
)PARTITION BY LIST(partition_id);

CREATE TABLE message_0 PARTITION OF message_new FOR VALUES IN (0);

-- Copy data from the old table to the new partitioned table
INSERT INTO message_new (transaction_hash, index, type, value, involved_accounts_addresses, partition_id, height)
SELECT message.transaction_hash, message.index, message.type, message.value, message.involved_accounts_addresses, 0, transaction.height
FROM message
LEFT JOIN transaction on message.transaction_hash = transaction.hash;

-- Rename the tables
ALTER TABLE message RENAME TO message_old;
ALTER TABLE message_new RENAME TO message;

-- DROP REFERENCES
DROP FUNCTION messages_by_address;
DROP FUNCTION gravity_messages_by_address;

-- Drop the old table
DROP TABLE message_old;

-- Create indexes on the new table
CREATE INDEX message_transaction_hash_index ON message (transaction_hash);
CREATE INDEX message_type_index ON message (type);
CREATE INDEX message_involved_accounts_index ON message USING GIN(involved_accounts_addresses);

-- MESSAGES_BY_ADDRESS
CREATE FUNCTION messages_by_address(
    addresses TEXT[],
    types TEXT[],
    "limit" BIGINT = 100,
    "offset" BIGINT = 0)
    RETURNS SETOF message AS
$$
SELECT * FROM message
WHERE (cardinality(types) = 0 OR type = ANY (types))
  AND addresses && involved_accounts_addresses
ORDER BY height DESC LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;


-- 04-staking.sql
ALTER TABLE staking_pool ADD COLUMN unbonding_tokens TEXT NOT NULL DEFAULT '0';
ALTER TABLE staking_pool ADD COLUMN staked_not_bonded_tokens TEXT NOT NULL DEFAULT '0';

ALTER TABLE validator_status DROP COLUMN tombstoned;

-- 09-gov.sql
ALTER TABLE gov_params ADD COLUMN params JSONB;
UPDATE gov_params SET params = deposit_params || voting_params || tally_params || '{"burn_vote_veto": true, "min_initial_deposit_ratio": "0.000000000000000000"}';
ALTER TABLE gov_params ALTER COLUMN params SET NOT NULL;

ALTER TABLE gov_params DROP COLUMN deposit_params;
ALTER TABLE gov_params DROP COLUMN voting_params;
ALTER TABLE gov_params DROP COLUMN tally_params;

ALTER TABLE proposal ADD COLUMN metadata TEXT NOT NULL DEFAULT '';
ALTER TABLE proposal DROP COLUMN proposal_route;
ALTER TABLE proposal DROP COLUMN proposal_type;

ALTER TABLE proposal_deposit ADD COLUMN timestamp TIMESTAMP;

ALTER TABLE proposal_vote ADD COLUMN timestamp TIMESTAMP;

ALTER TABLE proposal_staking_pool_snapshot
ALTER COLUMN bonded_tokens TYPE TEXT;
ALTER TABLE proposal_staking_pool_snapshot
ALTER COLUMN bonded_tokens SET NOT NULL;

ALTER TABLE proposal_staking_pool_snapshot
ALTER COLUMN not_bonded_tokens TYPE TEXT;
ALTER TABLE proposal_staking_pool_snapshot
ALTER COLUMN not_bonded_tokens SET NOT NULL;

-- 13-upgrade
CREATE TABLE software_upgrade_plan
(
    proposal_id     INTEGER REFERENCES proposal (id) UNIQUE,
    plan_name       TEXT        NOT NULL,
    upgrade_height  BIGINT      NOT NULL,
    info            TEXT        NOT NULL,
    height          BIGINT      NOT NULL
);
CREATE INDEX software_upgrade_plan_proposal_id_index ON software_upgrade_plan (proposal_id);
CREATE INDEX software_upgrade_plan_height_index ON software_upgrade_plan (height);

-- DISTINCT MESSAGE
ALTER TABLE distinct_message ADD partition_id BIGINT NULL;

ALTER TABLE distinct_message 
ADD CONSTRAINT distinct_message_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

DROP function messages_by_address_distinct_on_tx_hash;
CREATE FUNCTION messages_by_address_distinct_on_tx_hash(
  addresses TEXT[],
  types TEXT[],
  "limit" BIGINT = 100,
  "offset" BIGINT = 0)
RETURNS SETOF distinct_message AS
$$
SELECT DISTINCT ON(message.height, message.transaction_hash) message.transaction_hash, message.height, message.index, message.type, message.value, message.involved_accounts_addresses, message.partition_id
FROM message
   JOIN transaction t on message.transaction_hash = t.hash
WHERE (cardinality(types) = 0 OR type = ANY (types))
   AND addresses && involved_accounts_addresses
ORDER BY message.height DESC, message.transaction_hash DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;

-- GROUP MODULE
ALTER TABLE group_proposal ADD partition_id BIGINT NULL;

ALTER TABLE group_proposal 
ADD CONSTRAINT group_proposal_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

-- NFT MODULE
ALTER TABLE nft_denom ADD partition_id BIGINT NULL;

ALTER TABLE nft_denom 
ADD CONSTRAINT nft_denom_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE nft_nft ADD partition_id BIGINT NULL;

ALTER TABLE nft_nft 
ADD CONSTRAINT nft_nft_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

-- MARKETPLACE MODULE
ALTER TABLE marketplace_collection ADD partition_id BIGINT NULL;

ALTER TABLE marketplace_collection 
ADD CONSTRAINT marketplace_collection_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE marketplace_nft ADD partition_id BIGINT NULL;

ALTER TABLE marketplace_nft 
ADD CONSTRAINT marketplace_nft_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE marketplace_nft_buy_history ADD partition_id BIGINT NULL;

ALTER TABLE marketplace_nft_buy_history 
ADD CONSTRAINT marketplace_nft_buy_history_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE nft_transfer_history ADD partition_id BIGINT NULL;

ALTER TABLE nft_transfer_history 
ADD CONSTRAINT nft_transfer_history_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

-- COSM WASM
ALTER TABLE cosmwasm_store ADD partition_id BIGINT NULL;

ALTER TABLE cosmwasm_store 
ADD CONSTRAINT cosmwasm_store_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE cosmwasm_instantiate ADD partition_id BIGINT NULL;

ALTER TABLE cosmwasm_instantiate 
ADD CONSTRAINT cosmwasm_instantiate_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE cosmwasm_execute ADD partition_id BIGINT NULL;

ALTER TABLE cosmwasm_execute 
ADD CONSTRAINT cosmwasm_execute_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE cosmwasm_migrate ADD partition_id BIGINT NULL;

ALTER TABLE cosmwasm_migrate 
ADD CONSTRAINT cosmwasm_migrate_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE cosmwasm_update_admin ADD partition_id BIGINT NULL;

ALTER TABLE cosmwasm_update_admin 
ADD CONSTRAINT cosmwasm_update_admin_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

ALTER TABLE cosmwasm_clear_admin ADD partition_id BIGINT NULL;

ALTER TABLE cosmwasm_clear_admin 
ADD CONSTRAINT cosmwasm_clear_admin_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);

-- GRAVITY
ALTER TABLE gravity_transaction ADD partition_id BIGINT NULL;

ALTER TABLE gravity_transaction 
ADD CONSTRAINT gravity_transaction_transaction_hash_partition_id_fkey 
FOREIGN KEY (transaction_hash, partition_id) 
REFERENCES transaction (hash, partition_id);


CREATE FUNCTION gravity_messages_by_address(
    receiver_addr TEXT,
    "limit" BIGINT = 100,
    "offset" BIGINT = 0)
    RETURNS SETOF message AS
$$
SELECT m.transaction_hash, m.index, m.type, m.value, m.involved_accounts_addresses, m.partition_id, m.height
FROM message m
         JOIN gravity_transaction t on m.transaction_hash = t.transaction_hash
WHERE t.receiver = receiver_addr AND t.orchestrator = ANY(m.involved_accounts_addresses)
ORDER BY t.height DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;

