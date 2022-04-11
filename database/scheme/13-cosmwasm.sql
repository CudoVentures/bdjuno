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