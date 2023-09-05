UPDATE nft_nft
SET uniq_id = concat(id, '@', denom_id);

UPDATE nft_transfer_history
SET uniq_id = concat(id, '@', denom_id);

UPDATE marketplace_nft
SET uniq_id = concat(token_id, '@', denom_id);


UPDATE marketplace_nft_buy_history
SET uniq_id = concat(token_id, '@', denom_id);