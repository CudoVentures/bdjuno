package test

import (
	"fmt"
	"testing"

	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/forbole/bdjuno/v4/integration_tests/types"
	"github.com/stretchr/testify/require"
)

var (
	marketplaceModule      = "marketplace"
	marketplaceNftName     = "marketplaceNftName"
	marketplaceNftNDenomID = "marketplacenftnftdenomid"
	marketplaceNftSymbol   = "MKTNFTSMBL"
)

// Prerequisite for other tests
func addAdmin(t *testing.T, adminAddress string) {
	addAdminArgs := []string{
		marketplaceModule,
		"add-admin",
		adminAddress,
	}
	result, err := config.ExecuteTxCommand(CudosAdmin, addAdminArgs...)
	require.NoError(t, err)

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)
}

func TestCreateCollection(t *testing.T) {
	// PREPARE
	addAdmin(t, User1)
	verified := true
	args := []string{
		marketplaceModule,
		"create-collection",
		marketplaceNftNDenomID,
		config.GetFlag(Name, marketplaceNftName),
		config.GetFlag(Symbol, marketplaceNftSymbol),
		config.GetFlag(Schema, Schema),
		config.GetFlag(Traits, NotEditable),
		config.GetFlag(Description, Description),
		config.GetFlag(Minter, User1),
		config.GetFlag(Data, Data),
		config.GetFlag(MintRoyalties, Royalties),
		config.GetFlag(ResaleRoyalties, Royalties),
		fmt.Sprintf("--verified=%v", verified),
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	expectedCollection := types.MarketplaceCollectionQuery{
		DenomID:  marketplaceNftNDenomID,
		Verified: verified,
		Creator:  User1,
	}

	var collectionFromDB types.MarketplaceCollectionQuery
	err = config.QueryDatabase(`SELECT denom_id, mint_royalties, resale_royalties, verified, creator FROM marketplace_collection WHERE transaction_hash = $1`, txHash).Scan(
		&collectionFromDB.DenomID,
		&collectionFromDB.MintRoyalties,
		&collectionFromDB.ResaleRoyalties,
		&collectionFromDB.Verified,
		&collectionFromDB.Creator,
	)
	require.NoError(t, err)
	require.Equal(t, expectedCollection.DenomID, collectionFromDB.DenomID)
	require.Equal(t, expectedCollection.Verified, collectionFromDB.Verified)
	require.Equal(t, expectedCollection.Creator, collectionFromDB.Creator)
	require.NotEmpty(t, collectionFromDB.MintRoyalties)
	require.NotEmpty(t, collectionFromDB.ResaleRoyalties)
}

func TestPublishCollection(t *testing.T) {
	//PREPARE
	// Dependant on existing nft denom
	args := []string{
		marketplaceModule,
		"publish-collection",
		NftDenomId,
		config.GetFlag(MintRoyalties, Royalties),
		config.GetFlag(ResaleRoyalties, Royalties),
	}
	// EXECUTE
	result, err := config.ExecuteTxCommand(User1, args...)
	require.NoError(t, err)

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)
}
