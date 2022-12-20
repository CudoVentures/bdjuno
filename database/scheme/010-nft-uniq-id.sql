ALTER TABLE marketplace_nft_buy_history ADD uniq_id TEXT DEFAULT '';
ALTER TABLE nft_transfer_history ADD uniq_id TEXT DEFAULT '';
ALTER TABLE nft_nft ADD uniq_id TEXT DEFAULT '';
ALTER TABLE marketplace_nft ADD uniq_id TEXT DEFAULT '';

CREATE INDEX marketplace_nft_buy_history_uniq_id_index ON marketplace_nft_buy_history (uniq_id);
CREATE INDEX nft_transfer_history_uniq_id_index ON nft_transfer_history (uniq_id);
CREATE INDEX nft_nft_uniq_id_index ON nft_nft (uniq_id);
CREATE INDEX marketplace_nft_uniq_id_index ON marketplace_nft (uniq_id);