
CREATE TABLE nft_mint
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    token_id BIGINT NOT NULL,
    denom_id TEXT NOT NULL,
    PRIMARY KEY(token_id, denom_id)
);
