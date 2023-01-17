package handlers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
	nftTypes "github.com/CudoVentures/cudos-node/x/nft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	actionstypes "github.com/forbole/bdjuno/v2/cmd/actions/types"
	"github.com/forbole/bdjuno/v2/utils"
	"github.com/forbole/juno/v2/types"
	tendermintTypes "github.com/tendermint/tendermint/abci/types"
)

func NftTransferEvents(ctx *actionstypes.Context, payload *actionstypes.NftTransferEventsPayload) (interface{}, error) {
	if err := validatePayload(payload); err != nil {
		return nil, err
	}

	transferEvents := []actionstypes.TransferEvent{}

	transfersQuery := fmt.Sprintf("transfer_nft.token_id=%d AND transfer_nft.denom_id='%s'", payload.Input.TokenID, payload.Input.DenomID)
	junoTxs, err := searchTxsByFilter(transfersQuery, ctx)
	if err != nil {
		return nil, err
	}

	mintQuery := fmt.Sprintf("marketplace_mint_nft.token_id=%d AND marketplace_mint_nft.denom_id='%s'", payload.Input.TokenID, payload.Input.DenomID)
	mintJunoTxs, err := searchTxsByFilter(mintQuery, ctx)
	if err != nil {
		return nil, err
	}

	buyQuery := fmt.Sprintf("buy_nft.token_id=%d AND buy_nft.denom_id='%s'", payload.Input.TokenID, payload.Input.DenomID)
	buyJunoTxs, err := searchTxsByFilter(buyQuery, ctx)
	if err != nil {
		return nil, err
	}

	junoTxs = append(junoTxs, mintJunoTxs...)
	junoTxs = append(junoTxs, buyJunoTxs...)

	if payload.Input.FromTime != 0 {
		var err error
		junoTxs, err = filterTxsByTimeRange(junoTxs, payload.Input.FromTime, payload.Input.ToTime)
		if err != nil {
			return nil, err
		}
	}

	for _, junoTx := range junoTxs {
		txEvents, err := getTxNftChangeOwnershipEvents(ctx.Cdc, junoTx)
		if err != nil {
			return nil, err
		}

		transferEvents = append(transferEvents, txEvents...)
	}

	return actionstypes.TransferEventsResponse{TransferEvents: transferEvents}, nil
}

func validatePayload(payload *actionstypes.NftTransferEventsPayload) error {
	if (payload.Input.FromTime == 0 && payload.Input.ToTime != 0) ||
		(payload.Input.FromTime != 0 && payload.Input.ToTime == 0) {
		return errors.New("both from_time and to_time must be set")
	}
	return nil
}

func searchTxsByFilter(query string, ctx *actionstypes.Context) ([]*types.Tx, error) {
	var page = 1
	var perPage = 100
	var stop = false

	junoTxs := []*types.Tx{}

	for !stop {
		result, err := ctx.Node.TxSearch(query, &page, &perPage, "")
		if err != nil {
			return nil, fmt.Errorf("error while running tx search: %s", err)
		}

		for _, tx := range result.Txs {
			junoTx, err := ctx.Node.Tx(hex.EncodeToString(tx.Tx.Hash()))
			if err != nil {
				return nil, err
			}

			junoTxs = append(junoTxs, junoTx)
		}

		page++

		stop = len(junoTxs) == result.TotalCount
	}

	return junoTxs, nil
}

func getTxNftChangeOwnershipEvents(cdc codec.Codec, tx *types.Tx) ([]actionstypes.TransferEvent, error) {
	transferEvents := []actionstypes.TransferEvent{}

	timestamp, err := utils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return nil, err
	}

	duplicatedEvents := make(map[string]bool)

	appendIfNotDuplicated := func(transferEvent actionstypes.TransferEvent) {
		key := fmt.Sprintf("%v", transferEvent)
		if ok := duplicatedEvents[key]; !ok {
			duplicatedEvents[key] = true
			transferEvents = append(transferEvents, transferEvent)
		}
	}

	for _, msg := range tx.Body.Messages {
		var stdMsg sdk.Msg
		if err := cdc.UnpackAny(msg, &stdMsg); err != nil {
			return nil, fmt.Errorf("error while unpacking message: %s", err)
		}

		switch cosmosMsg := stdMsg.(type) {
		case *marketplaceTypes.MsgMintNft:
			appendIfNotDuplicated(actionstypes.TransferEvent{
				From:      "0x0",
				To:        cosmosMsg.Recipient,
				Timestamp: timestamp,
			})

		case *marketplaceTypes.MsgBuyNft:
			{
				events := tx.Events
				for _, event := range events {
					if event.Type == marketplaceTypes.EventBuyNftType {
						from := getAttributeValueFromEvent(event, marketplaceTypes.AttributeKeyOwner)
						to := getAttributeValueFromEvent(event, marketplaceTypes.AttributeKeyBuyer)
						appendIfNotDuplicated(actionstypes.TransferEvent{
							From:      from,
							To:        to,
							Timestamp: timestamp,
						})
						break
					}
				}
			}
		case *nftTypes.MsgTransferNft:
			appendIfNotDuplicated(actionstypes.TransferEvent{
				From:      cosmosMsg.From,
				To:        cosmosMsg.To,
				Timestamp: timestamp,
			})
		}
	}

	if len(transferEvents) == 0 {
		return nil, errors.New("failed to parse change ownership events")
	}

	return transferEvents, nil
}

// TODO: Can be optimized by looking up closes height by time from DB and
// including it in TxSearch filter or add timestamp to mint and transfer events
// https://docs.tendermint.com/v0.34/app-dev/indexing-transactions.html

func filterTxsByTimeRange(junoTxs []*types.Tx, from, to int) ([]*types.Tx, error) {
	filtered := []*types.Tx{}
	for _, junoTx := range junoTxs {
		timestamp, err := utils.ISO8601ToTimestamp(junoTx.Timestamp)
		if err != nil {
			return nil, err
		}

		if timestamp > from && timestamp <= to {
			filtered = append(filtered, junoTx)
		}
	}
	return filtered, nil
}

func getAttributeValueFromEvent(event tendermintTypes.Event, attributeKey string) string {
	for _, attr := range event.Attributes {
		if string(attr.Key) == attributeKey {
			return strings.ReplaceAll(string(attr.Value), "\"", "")
		}
	}
}
