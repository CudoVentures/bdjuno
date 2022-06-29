/* ---- APR ---- */

CREATE TABLE apr
(
    one_row_id bool PRIMARY KEY DEFAULT TRUE,
    value      DECIMAL NOT NULL,
    height     BIGINT  NOT NULL,
    CONSTRAINT one_row_uni CHECK (one_row_id)
);

CREATE INDEX apr_height_index ON apr (height);

CREATE TABLE apr_history
(
    value DECIMAL NOT NULL,
    height BIGINT  NOT NULL,
    timestamp BIGINT NOT NULL
);

CREATE INDEX apr_history_height_index ON apr_history (height);
CREATE INDEX apr_history_timestamp_index ON apr_history (timestamp);

/* --- Adjusted Supply --- */

CREATE TABLE adjusted_supply
(
    one_row_id bool PRIMARY KEY DEFAULT TRUE,
    value      DECIMAL NOT NULL,
    height     BIGINT  NOT NULL,
    CONSTRAINT one_row_uni CHECK (one_row_id)
);
