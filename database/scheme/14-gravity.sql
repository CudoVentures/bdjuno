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
    PRIMARY KEY(attestation_id, orchestrator)
);

CREATE INDEX gravity_transaction_receiver_index ON gravity_transaction (receiver);
