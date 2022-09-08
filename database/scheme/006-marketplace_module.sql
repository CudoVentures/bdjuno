
/* Extend NFT denom table with the newly added columns for the marketplace use cases */

ALTER TABLE marketplace ADD traits TEXT DEFAULT '';
ALTER TABLE marketplace ADD minter TEXT DEFAULT '';
ALTER TABLE marketplace ADD description TEXT DEFAULT '';

/* Marketplace entities */

CREATE TABLE marketplace_collection
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    id BIGINT NOT NULL,
    denom_id TEXT NOT NULL REFERENCES nft_denom (id),
    mint_royalties TEXT NOT NULL,
    resale_royalties TEXT NOT NULL,
    creator TEXT NOT NULL,
    PRIMARY KEY(id)
);

CREATE INDEX marketplace_collection_denom_id_index ON marketplace_collection (denom_id);
CREATE INDEX marketplace_collection_creator_index ON marketplace_collection (creator);

CREATE TABLE marketplace_nft
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    id BIGINT NOT NULL,
    token_id BIGINT NOT NULL REFERENCES nft_nft (id),
    denom_id TEXT NOT NULL REFERENCES nft_denom (id),
    price TEXT NOT NULL,
    creator TEXT NOT NULL,
    PRIMARY KEY(id)
);

CREATE INDEX marketplace_nft_collection_id_index ON marketplace_nft (collection_id);
CREATE INDEX marketplace_nft_token_id_denom_id_index ON marketplace_nft (token_id, denom_id);
CREATE INDEX marketplace_nft_creator_index ON marketplace_nft (creator);

CREATE TABLE marketplace_nft_buy_history
(
    transaction_hash TEXT NOT NULL REFERENCES transaction (hash),
    token_id BIGINT NOT NULL REFERENCES nft_nft (id),
    denom_id TEXT NOT NULL REFERENCES nft_denom (id),
    price TEXT NOT NULL,
    seller TEXT NOT NULL,
    buyer TEXT NOT NULL,
    timestamp BIGINT NOT NULL,
)

CREATE INDEX marketplace_nft_buy_history_token_id_denom_id_index ON marketplace_nft_buy_history (token_id, denom_id);
CREATE INDEX marketplace_nft_buy_history_seller_index ON marketplace_nft_buy_history (seller);
CREATE INDEX marketplace_nft_buy_history_buyer_index ON marketplace_nft_buy_history (buyer);
CREATE INDEX marketplace_nft_buy_history_timestamp_index ON marketplace_nft_buy_history (timestamp);