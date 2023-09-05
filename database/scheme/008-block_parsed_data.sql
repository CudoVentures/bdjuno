CREATE TABLE block_parsed_data
(
    height            BIGINT  PRIMARY KEY,
    validators        BOOLEAN NOT NULL,
    block             BOOLEAN NOT NULL,
    commits           BOOLEAN NOT NULL,
    txs               TEXT    NOT NULL,
    all_txs           BOOLEAN NOT NULL,
    block_modules     TEXT    NOT NULL,
    all_block_modules BOOLEAN NOT NULL
)
