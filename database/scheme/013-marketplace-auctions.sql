CREATE TABLE marketplace_auction (
    id           BIGINT                      NOT NULL PRIMARY KEY,
    token_id     BIGINT                      NOT NULL,
    denom_id     TEXT                        NOT NULL REFERENCES nft_denom (id),
    creator      TEXT                        NOT NULL,
    start_time   TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    end_time     TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    auction      TEXT                        NOT NULL,
    sold         BOOLEAN                     NOT NULL DEFAULT FALSE,
    FOREIGN KEY (token_id, denom_id) REFERENCES nft_nft(id, denom_id)
);

CREATE INDEX marketplace_auction_token_id_denom_id_index ON marketplace_auction (token_id, denom_id);
CREATE INDEX marketplace_auction_creator_index ON marketplace_auction (creator);

CREATE TABLE marketplace_bid (
    auction_id BIGINT                      NOT NULL REFERENCES marketplace_auction (id),
    bidder     TEXT                        NOT NULL,
    price      DECIMAL                     NOT NULL,
    timestamp  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    transaction_hash TEXT                  NOT NULL REFERENCES transaction (hash)
);

CREATE INDEX marketplace_bid_auction_id_index ON marketplace_bid (auction_id);
CREATE INDEX marketplace_bid_bidder_index ON marketplace_bid (bidder);
CREATE INDEX marketplace_bid_timestamp_index ON marketplace_bid (timestamp desc);
