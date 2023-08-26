
CREATE TABLE nft_denom
(
    transaction_hash TEXT NOT NULL,
    partition_id BIGINT NULL,
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    schema TEXT NOT NULL,
    symbol TEXT NOT NULL,
    owner TEXT NOT NULL,
    contract_address_signer TEXT NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY(transaction_hash, partition_id) REFERENCES transaction (hash, partition_id)
);

CREATE INDEX nft_denom_owner_index ON nft_denom (owner);

CREATE TABLE nft_nft
(
    transaction_hash TEXT NOT NULL,
    partition_id BIGINT NULL,
    id BIGINT NOT NULL,
    denom_id TEXT NOT NULL REFERENCES nft_denom (id),
    name TEXT NOT NULL,
    uri TEXT NOT NULL,
    data_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    data_text TEXT NOT NULL DEFAULT '',
    owner TEXT NOT NULL,
    sender TEXT NOT NULL,
    contract_address_signer TEXT NOT NULL,
    burned BOOLEAN DEFAULT FALSE,
    PRIMARY KEY(id, denom_id),
    FOREIGN KEY(transaction_hash, partition_id) REFERENCES transaction (hash, partition_id)
);

CREATE INDEX nft_nft_owner_index ON nft_nft (owner);

CREATE FUNCTION nfts_by_data_property(
     property_name TEXT,
     "limit" BIGINT = 100,
     "offset" BIGINT = 0)
     RETURNS SETOF nft_nft AS
   $$
SELECT * FROM nft_nft WHERE data_json[property_name] IS NOT NULL LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;

CREATE FUNCTION nfts_by_expiration_date(
     expiration_date BIGINT,
     "limit" BIGINT = 100,
     "offset" BIGINT = 0)
     RETURNS SETOF nft_nft AS
   $$
SELECT * FROM nft_nft WHERE data_json['expirationDate'] IS NOT NULL AND data_json['expirationDate']::bigint >= expiration_date LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;