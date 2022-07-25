
CREATE TABLE nft_issue_denom
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    denom_id TEXT NOT NULL,
    PRIMARY KEY(denom_id)
);
