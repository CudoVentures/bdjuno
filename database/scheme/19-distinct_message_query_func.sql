CREATE TABLE distinct_message
(	
  transaction_hash TEXT NOT NULL,
  height BIGINT,
  index                       BIGINT,
  type                        TEXT,
  value                       JSONB,
  involved_accounts_addresses TEXT[],
  partition_id BIGINT NULL,
  FOREIGN KEY(transaction_hash, partition_id) REFERENCES transaction (hash, partition_id)
);

CREATE FUNCTION messages_by_address_distinct_on_tx_hash(
  addresses TEXT[],
  types TEXT[],
  "limit" BIGINT = 100,
  "offset" BIGINT = 0)
RETURNS SETOF distinct_message AS
$$
SELECT DISTINCT ON(message.height, message.transaction_hash) message.transaction_hash, message.height, message.index, message.type, message.value, message.involved_accounts_addresses, message.partition_id
FROM message
   JOIN transaction t on message.transaction_hash = t.hash
WHERE (cardinality(types) = 0 OR type = ANY (types))
   AND addresses && involved_accounts_addresses
ORDER BY message.height DESC, message.transaction_hash DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;