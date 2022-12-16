CREATE TABLE cw20token_allowance
(
    token   TEXT NOT NULL REFERENCES cw20token_info(address) ON DELETE CASCADE,
    owner   TEXT NOT NULL,
    spender TEXT NOT NULL,
    amount  TEXT NOT NULL,
    expires TEXT NULL,
    PRIMARY KEY (token, owner, spender)
);

CREATE INDEX cw20token_allowance_owner_index ON cw20token_allowance (owner);
