package test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/forbole/bdjuno/v4/integration_tests/types"
	"github.com/stretchr/testify/require"
)

var (
	marketplaceModule                = "marketplace"
	marketplaceNftName               = "marketplaceNftName"
	marketplaceNftNDenomID           = "marketplacenftnftdenomid"
	marketplaceNftSymbol             = "MKTNFTSMBL"
	marketplaceUnverifiedNftName     = "marketplaceUnverifiedNftName"
	marketplaceUnverifiedNftNDenomID = "marketplaceunverifiednftnftdenomid"
	marketplaceUnverifiedNftSymbol   = "unverifiedMKTNFTSMBL"
	expectedExistingID               = "1"
	expectedUniqID                   = fmt.Sprintf("%s@%s", expectedExistingID, marketplaceNftNDenomID)
)

// Prerequisite for other tests

func getMarketplaceNftPrice(t *testing.T, uniqID string) string {
	var price string
	err := config.QueryDatabase(`SELECT price FROM marketplace_nft WHERE uniq_id = $1`, uniqID).Scan(
		&price,
	)
	require.NoError(t, err)
	return price
}

func getCollectionRoyaltiesAsCLIString(t *testing.T, denomID string) (string, string) {
	var mintRoyalties string
	var resaleRoyalties string

	err := config.QueryDatabase(`SELECT mint_royalties, resale_royalties FROM marketplace_collection WHERE denom_id = $1`, denomID).Scan(
		&mintRoyalties,
		&resaleRoyalties,
	)
	require.NoError(t, err)

	var typedMintroyalties []types.Royalty
	var typedResaleroyalties []types.Royalty
	err = json.Unmarshal([]byte(mintRoyalties), &typedMintroyalties)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resaleRoyalties), &typedResaleroyalties)
	require.NoError(t, err)

	var mintRecords []string
	for _, ap := range typedMintroyalties {
		floatValue, err := strconv.ParseFloat(ap.Percent, 64)
		require.NoError(t, err)
		intValue := int(floatValue)
		intStr := strconv.Itoa(intValue)
		mintRecords = append(mintRecords, fmt.Sprintf("%s:%s", ap.Address, intStr))
	}

	var resaleRecords []string
	for _, ap := range typedResaleroyalties {
		floatValue, err := strconv.ParseFloat(ap.Percent, 64)
		require.NoError(t, err)
		intValue := int(floatValue)
		intStr := strconv.Itoa(intValue)
		resaleRecords = append(resaleRecords, fmt.Sprintf("%s:%s", ap.Address, intStr))
	}

	return strings.Join(mintRecords, ","), strings.Join(resaleRecords, ",")
}

func getCollectionID(t *testing.T, denomID string) string {
	var ID int
	err := config.QueryDatabase(`SELECT id FROM marketplace_collection WHERE denom_id = $1`, denomID).Scan(
		&ID,
	)
	require.NoError(t, err)
	return fmt.Sprintf("%d", ID)
}

func getNftID(t *testing.T, uniqID string) string {
	var ID sql.NullInt64
	err := config.QueryDatabase(`SELECT id FROM marketplace_nft WHERE uniq_id = $1`, uniqID).Scan(
		&ID,
	)
	require.NoError(t, err)
	return fmt.Sprintf("%v", ID.Int64)
}

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

func TestCreateVerifiedCollection(t *testing.T) {
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

func TestCreateUnverifiedCollection(t *testing.T) {

	// PREPARE
	args := []string{
		marketplaceModule,
		"create-collection",
		marketplaceUnverifiedNftNDenomID,
		config.GetFlag(Name, marketplaceUnverifiedNftName),
		config.GetFlag(Symbol, marketplaceUnverifiedNftSymbol),
		config.GetFlag(Schema, Schema),
		config.GetFlag(Traits, NotEditable),
		config.GetFlag(Description, Description),
		config.GetFlag(Minter, User1),
		config.GetFlag(Data, Data),
		config.GetFlag(MintRoyalties, Royalties),
		config.GetFlag(ResaleRoyalties, Royalties),
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

	expectedVerifiedStatus := false

	var statusFromDB bool
	err = config.QueryDatabase(`SELECT verified FROM marketplace_collection WHERE denom_id = $1`, marketplaceUnverifiedNftNDenomID).Scan(
		&statusFromDB,
	)
	require.NoError(t, err)
	require.Equal(t, expectedVerifiedStatus, statusFromDB)
}

func TestVerifyAndUnverifyCollection(t *testing.T) {
	// Dependand on existing admin
	// PREPARE
	collectionID := getCollectionID(t, marketplaceUnverifiedNftNDenomID)
	verifyArgs := []string{
		marketplaceModule,
		"verify-collection",
		collectionID,
	}
	unverifyArgs := []string{
		marketplaceModule,
		"unverify-collection",
		collectionID,
	}

	// EXECUTE VERIFY
	result, err := config.ExecuteTxCommand(User1, verifyArgs...)
	require.NoError(t, err)

	// ASSERT VERIFIED

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	expectedVerifiedStatus := true

	var statusFromDB bool
	err = config.QueryDatabase(`SELECT verified FROM marketplace_collection WHERE denom_id = $1`, marketplaceUnverifiedNftNDenomID).Scan(
		&statusFromDB,
	)
	require.NoError(t, err)
	require.Equal(t, expectedVerifiedStatus, statusFromDB)

	// EXECUTE UNVERIFY
	result, err = config.ExecuteTxCommand(User1, unverifyArgs...)
	require.NoError(t, err)

	// ASSERT UNVERIFIED

	// make sure TX is included on chain
	txHash, blockHeight, err = config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists = config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	expectedVerifiedStatus = false

	err = config.QueryDatabase(`SELECT verified FROM marketplace_collection WHERE denom_id = $1`, marketplaceUnverifiedNftNDenomID).Scan(
		&statusFromDB,
	)
	require.NoError(t, err)
	require.Equal(t, expectedVerifiedStatus, statusFromDB)
}

func TestUpdateRoyalties(t *testing.T) {
	// PREPARE
	currentMintRoyalties, currentResaleRoyalties := getCollectionRoyaltiesAsCLIString(t, marketplaceNftNDenomID)
	require.Equal(t, Royalties, currentMintRoyalties)
	require.Equal(t, Royalties, currentResaleRoyalties)

	collectionID := getCollectionID(t, marketplaceNftNDenomID)
	args := []string{
		marketplaceModule,
		"update-royalties",
		collectionID,
		config.GetFlag(MintRoyalties, UpdatedRoyalties),
		config.GetFlag(ResaleRoyalties, UpdatedRoyalties),
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

	updatedMintRoyalties, updatedResaleRoyalties := getCollectionRoyaltiesAsCLIString(t, marketplaceNftNDenomID)
	require.Equal(t, UpdatedRoyalties, updatedMintRoyalties)
	require.Equal(t, UpdatedRoyalties, updatedResaleRoyalties)
}

func TestPublishCollection(t *testing.T) {
	// PREPARE
	// Dependent on existing nft denom, which is not published as marketplace collection yet!!!
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

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	err = config.QueryDatabase(`
	SELECT EXISTS(
		SELECT 1 FROM marketplace_collection 
		WHERE transaction_hash = $1 
		AND denom_id = $2 
		AND creator = $3
	)`,
		txHash, NftDenomId, User1,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestMintMarketplaceNftToCollection(t *testing.T) {

	// PREPARE
	recipient := User1
	args := []string{
		marketplaceModule,
		"mint-nft",
		marketplaceNftNDenomID,
		recipient,
		NftPrice,
		marketplaceNftName,
		config.GetFlag(URI, URI),
		config.GetFlag(Data, Data),
		config.GetFlag(UID, UID),
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(recipient, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// Marketplace is supposed to write to 4 different tables at this point, so we have to make sure it really did.

	// nft_nft
	err = config.QueryDatabase(`
		SELECT EXISTS(
			SELECT 1 FROM nft_nft 
			WHERE transaction_hash = $1 
			AND denom_id = $2 
			AND name = $3 
			AND owner = $4
		)`,
		txHash, marketplaceNftNDenomID, marketplaceNftName, recipient,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

	// marketplace_nft_buy_history
	err = config.QueryDatabase(`
		SELECT EXISTS(
			SELECT 1 FROM marketplace_nft_buy_history 
			WHERE transaction_hash = $1 
			AND denom_id = $2 
			AND buyer = $3
		)`,
		txHash, marketplaceNftNDenomID, recipient,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

	// marketplace_nft
	err = config.QueryDatabase(`
		SELECT EXISTS(
			SELECT 1 FROM marketplace_nft 
			WHERE transaction_hash = $1 
			AND denom_id = $2 
			AND creator = $3
		)`,
		txHash, marketplaceNftNDenomID, recipient,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

	// nft_transfer_history
	err = config.QueryDatabase(`
	SELECT EXISTS(
		SELECT 1 FROM nft_transfer_history 
		WHERE transaction_hash = $1 
		AND denom_id = $2 
		AND new_owner = $3
	)`,
		txHash, marketplaceNftNDenomID, recipient,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestPublishMintedNftForSale(t *testing.T) {

	// Make sure the NFT is not on sale, thus current price = 0
	currentPrice := getMarketplaceNftPrice(t, expectedUniqID)
	require.Equal(t, "0", currentPrice)

	// PREPARE
	args := []string{
		marketplaceModule,
		"publish-nft",
		expectedExistingID,
		marketplaceNftNDenomID,
		NftPrice,
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

	// Make sure the NFT is now on sale, thus currentPrice = NftPrice
	currentPrice = getMarketplaceNftPrice(t, expectedUniqID)
	require.Contains(t, NftPrice, currentPrice)
}

func TestUpdateNftPrice(t *testing.T) {
	currentPrice := getMarketplaceNftPrice(t, expectedUniqID)
	intendedNewPrice := "123456789acudos"
	require.NotContains(t, intendedNewPrice, currentPrice)

	// PREPARE
	id := getNftID(t, expectedUniqID)
	args := []string{
		marketplaceModule,
		"update-price",
		id,
		intendedNewPrice,
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

	currentPrice = getMarketplaceNftPrice(t, expectedUniqID)
	require.Contains(t, intendedNewPrice, currentPrice)
}

func TestRemoveNftFromSale(t *testing.T) {
	currentPrice := getMarketplaceNftPrice(t, expectedUniqID)
	intendedNewPrice := "0"
	require.NotEqual(t, intendedNewPrice, currentPrice)

	// PREPARE
	id := getNftID(t, expectedUniqID)
	args := []string{
		marketplaceModule,
		"remove-nft",
		id,
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

	currentPrice = getMarketplaceNftPrice(t, expectedUniqID)
	require.Equal(t, intendedNewPrice, currentPrice)
}

func TestBuyNft(t *testing.T) {
	// PREPARE PUBLISH
	// First publish the previously removed NFT back to sale
	oldOwner := User1
	args := []string{
		marketplaceModule,
		"publish-nft",
		expectedExistingID,
		marketplaceNftNDenomID,
		NftPrice,
	}

	// EXECUTE PUBLISH
	result, err := config.ExecuteTxCommand(oldOwner, args...)
	require.NoError(t, err)

	// ASSERT PUBLISH

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// PREPARE BUY
	newOwner := User3
	id := getNftID(t, expectedUniqID)
	args = []string{
		marketplaceModule,
		"buy-nft",
		id,
	}

	// EXECUTE BUY
	result, err = config.ExecuteTxCommand(newOwner, args...)
	require.NoError(t, err)

	// ASSERT BUY

	// make sure TX is included on chain
	txHash, blockHeight, err = config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists = config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// marketplace_nft_buy_history
	var seller string
	var buyer string
	err = config.QueryDatabase(`SELECT seller, buyer FROM marketplace_nft_buy_history WHERE uniq_id = $1 ORDER BY timestamp DESC`, expectedUniqID).Scan(
		&seller,
		&buyer,
	)
	require.NoError(t, err)
	require.Equal(t, seller, oldOwner)
	require.Equal(t, buyer, newOwner)

	// nft_transfer_history
	err = config.QueryDatabase(`SELECT old_owner, new_owner FROM nft_transfer_history WHERE uniq_id = $1 ORDER BY timestamp DESC`, expectedUniqID).Scan(
		&seller,
		&buyer,
	)
	require.NoError(t, err)
	require.Equal(t, seller, oldOwner)
	require.Equal(t, buyer, newOwner)

	// Also to be removed from sale
	currentPrice := getMarketplaceNftPrice(t, expectedUniqID)
	require.Equal(t, "0", currentPrice)
}
