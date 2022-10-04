CREATE TABLE cw20token_code_id
(
    id INT NOT NULL PRIMARY KEY
);

CREATE TABLE cw20token_info
(
    address            TEXT   NOT NULL PRIMARY KEY,
    code_id            INT    NOT NULL REFERENCES cw20token_code_id(id),
    name               TEXT   NOT NULL,
    symbol             TEXT   NOT NULL,
    decimals           INT    NOT NULL,
    circulating_supply BIGINT NOT NULL,
    max_supply         BIGINT NULL,
    minter             TEXT   NULL,
    marketing_admin    TEXT   NULL,
    project_url        TEXT   NULL,
    description        TEXT   NULL,
    logo               JSONB  NULL
);

CREATE INDEX cw20token_info_code_id_index ON cw20token_info (code_id);


CREATE TABLE cw20token_balance
(
    address TEXT   NOT NULL,
    token   TEXT   NOT NULL REFERENCES cw20token_info(address),
    balance BIGINT NOT NULL,
    PRIMARY KEY (address, token)
);

CREATE INDEX cw20token_balance_token_index ON cw20token_balance (token);
CREATE INDEX cw20token_balance_address_index ON cw20token_balance (address);
