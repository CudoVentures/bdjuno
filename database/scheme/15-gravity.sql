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
    transaction_hash TEXT NOT NULL,
    partition_id BIGINT NOT NULL,
    height BIGINT NOT NULL REFERENCES block (height),
    PRIMARY KEY(attestation_id, orchestrator),
    FOREIGN KEY(transaction_hash, partition_id) REFERENCES transaction (hash, partition_id)
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
SELECT m.transaction_hash, m.index, m.type, m.value, m.involved_accounts_addresses, m.partition_id, m.height
FROM message m
         JOIN gravity_transaction t on m.transaction_hash = t.transaction_hash
WHERE t.receiver = receiver_addr AND t.orchestrator = ANY(m.involved_accounts_addresses)
ORDER BY t.height DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;
