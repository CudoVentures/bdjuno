package workers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/nft"
	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/rs/zerolog/log"
)

type migrateNftsWorker struct {
	baseWorker
}

func (mnw migrateNftsWorker) Name() string {
	return "migrate_nfts_worker"
}

func (mnw migrateNftsWorker) Start(ctx context.Context, parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage, interval time.Duration) {
	mnw.baseWorker.Start(ctx, mnw.Name(), mnw.migrateNfts, parseCfg, parseCtx, storage, interval)
}

func (mnw migrateNftsWorker) migrateNfts(parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage) error {
	currentHeightVal, err := storage.GetOrDefaultValue(nftMigrationCurrentHeightKey, "0")
	if err != nil {
		return fmt.Errorf("error while getting worker storage key '%s': %s", nftMigrationCurrentHeightKey, err)
	}

	currentHeight, err := strconv.ParseInt(currentHeightVal, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing current height string value '%s': %s", currentHeightVal, err)
	}

	if currentHeight == 0 {
		currentHeight, err = getGenesisMaxInitialHeight(parseCtx)
		if err != nil {
			return fmt.Errorf("error while getting genesis max initial height: %s", err)
		}
	}

	untilHeightVal, err := storage.GetOrDefaultValue(nftMigrateUntilHeightKey, "0")
	if err != nil {
		return fmt.Errorf("error while getting worker storage key '%s': %s", nftMigrateUntilHeightKey, err)
	}

	untilHeight, err := strconv.ParseInt(untilHeightVal, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing until height string value '%s': %s", untilHeightVal, err)
	}

	if untilHeight == 0 {
		untilHeight, err = parseCtx.Node.LatestHeight()
		if err != nil {
			return fmt.Errorf("error while getting chain latest block height: %s", err)
		}

		latestHeightVal := strconv.FormatInt(untilHeight, 10)

		if err := storage.SetValue(nftMigrateUntilHeightKey, latestHeightVal); err != nil {
			return fmt.Errorf("error while storing migrate until value '%s': %s", latestHeightVal, err)
		}
	}

	if currentHeight >= untilHeight {
		// Processing finished
		return nil
	}

	nftModule := nft.NewModule(parseCtx.EncodingConfig.Marshaler, database.Cast(parseCtx.Database))

	for i := 0; i < 10000; i++ {
		if err := mnw.processBlock(nftModule, parseCtx, currentHeight); err != nil {
			return fmt.Errorf("error while processing block at height '%d': %s", currentHeight, err)
		}

		currentHeight++
	}

	currentHeightVal = strconv.FormatInt(currentHeight, 10)
	if err := storage.SetValue(nftMigrationCurrentHeightKey, currentHeightVal); err != nil {
		return fmt.Errorf("error while storing migration current height value '%s': %s", currentHeightVal, err)
	}

	return nil
}

func (mnw migrateNftsWorker) processBlock(module *nft.Module, parseCtx *parse.Context, currentHeight int64) error {
	log.Debug().Str("worker", "migrate_nft_worker").Msg(fmt.Sprintf("Processing block at height %d", currentHeight))

	block, err := parseCtx.Node.Block(currentHeight)
	if err != nil {
		return fmt.Errorf("failed to get block from node: %s", err)
	}

	txs, err := parseCtx.Node.Txs(block)
	if err != nil {
		return fmt.Errorf("failed to get transactions for block: %s", err)
	}

	for _, tx := range txs {
		log.Debug().Str("worker", "migrate_nft_worker").Msg(fmt.Sprintf("Processing TX hash: %s at height %d", tx.TxHash, currentHeight))
		msgIndex := 0
		for _, msg := range tx.Body.Messages {
			var stdMsg sdk.Msg
			if err := parseCtx.EncodingConfig.Marshaler.UnpackAny(msg, &stdMsg); err != nil {
				return fmt.Errorf("error while unpacking message: %s", err)
			}

			if err := module.HandleMsg(msgIndex, stdMsg, tx); err != nil {
				return fmt.Errorf("error while nft module handle msg: %s", err)
			}

			msgIndex++
		}
	}

	return nil
}

const (
	nftMigrationCurrentHeightKey = "nft_migrate_current_height"
	nftMigrateUntilHeightKey     = "nft_migrate_until_height"
)
