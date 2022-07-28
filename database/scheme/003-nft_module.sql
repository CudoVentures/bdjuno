
CREATE TABLE nft_denom
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    schema TEXT NOT NULL,
    symbol TEXT NOT NULL,
    owner TEXT NOT NULL,
    contract_address_signer TEXT NOT NULL,
    PRIMARY KEY(id)
);

CREATE INDEX nft_denom_owner_index ON nft_denom (owner);

CREATE TABLE nft_nft
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    id BIGINT NOT NULL,
    denom_id TEXT NOT NULL REFERENCES nft_denom (denom_id),
    name TEXT NOT NULL,
    uri TEXT NOT NULL,
    data NOT NULL DEFAULT '[]'::JSONB,
    owner TEXT NOT NULL,
    sender TEXT NOT NULL,
    contract_address_signer TEXT NOT NULL,
    burned BOOLEAN DEFAULT FALSE,
    PRIMARY KEY(token_id, denom_id)
);

CREATE INDEX nft_nft_owner_index ON nft_nft (owner);

CREATE FUNCTION nfts_by_data_property(
     property_name TEXT,
     "limit" BIGINT = 100,
     "offset" BIGINT = 0)
     RETURNS SETOF nft_nft AS
   $$
SELECT * FROM nft_nft WHERE data[property_name] IS NOT NULL LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;