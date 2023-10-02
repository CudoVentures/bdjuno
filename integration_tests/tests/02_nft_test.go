package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/forbole/bdjuno/v4/integration_tests/types"
)

var (
	nftModule          = "nft"
	burnableNftName    = "burnableNftName"
	NftDenomId         = "nftdenomid"
	burnableNftDenomID = "burnablenftdenomid"
	burnableSymbol     = "burnableSymbol"
	nftID              = 1
	uniqBurnableNftId  = fmt.Sprintf("%d@%s", nftID, burnableNftDenomID)
)

func TestIssueNftDenom(t *testing.T) {

	// PREPARE
	args := []string{
		nftModule,
		"issue",
		NftDenomId,
		config.GetFlag(Name, Name),
		config.GetFlag(Symbol, Symbol),
		config.GetFlag(Schema, Schema),
		config.GetFlag(Description, Description),
		config.GetFlag(Data, Data),
		config.GetFlag(Traits, NotEditable),
		config.GetFlag(Minter, User1),
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	expectedDenom := types.NftDenomQuery{
		DenomID:     NftDenomId,
		Name:        Name,
		Schema:      Schema,
		Symbol:      Symbol,
		Owner:       User1,
		Traits:      NotEditable,
		Minter:      User1,
		Description: Description,
		DataText:    Data,
	}

	var denomFromDB types.NftDenomQuery
	err = config.QueryDatabase(`SELECT id, Name, Schema, Symbol, owner, Traits, Minter, Description, data_text FROM nft_denom where transaction_hash = $1`, txHash).Scan(
		&denomFromDB.DenomID,
		&denomFromDB.Name,
		&denomFromDB.Schema,
		&denomFromDB.Symbol,
		&denomFromDB.Owner,
		&denomFromDB.Traits,
		&denomFromDB.Minter,
		&denomFromDB.Description,
		&denomFromDB.DataText,
	)
	require.NoError(t, err)
	require.Equal(t, expectedDenom, denomFromDB)
}

func TestMintNftToDenom(t *testing.T) {
	// PREPARE BURNABLE DENOM so we can mint NFT to it
	argsForBurnableDenom := []string{
		nftModule,
		"issue",
		burnableNftDenomID,
		config.GetFlag(Name, burnableNftName),
		config.GetFlag(Symbol, burnableSymbol),
		config.GetFlag(Schema, Schema),
		config.GetFlag(Description, Description),
		config.GetFlag(Data, Data),
		config.GetFlag(Minter, User1),
	}
	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(User1, argsForBurnableDenom...)
	require.NoError(t, err)

	// ASSERT
	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// PREPARE MINT
	mintArgs := []string{
		nftModule,
		"mint",
		burnableNftDenomID,
		config.GetFlag(Recipient, User1),
		config.GetFlag(Name, burnableNftName),
		config.GetFlag(URI, URI),
	}

	// EXECUTE MINT
	result, err = config.ExecuteTxCommandWithFees(User1, mintArgs...)
	require.NoError(t, err)

	// ASSERT MINT
	// make sure TX is included on chain
	txHash, blockHeight, err = config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists = config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	expectedNft := types.NftQuery{
		ID:      nftID,
		DenomID: burnableNftDenomID,
		Name:    burnableNftName,
		URI:     URI,
		Owner:   User1,
		Sender:  User1,
		Burned:  false,
		UniqID:  uniqBurnableNftId,
	}

	var nftFromDB types.NftQuery
	err = config.QueryDatabase(`SELECT id, denom_id, Name, URI, owner, sender, burned, uniq_id FROM nft_nft WHERE transaction_hash = $1`, txHash).Scan(
		&nftFromDB.ID,
		&nftFromDB.DenomID,
		&nftFromDB.Name,
		&nftFromDB.URI,
		&nftFromDB.Owner,
		&nftFromDB.Sender,
		&nftFromDB.Burned,
		&nftFromDB.UniqID,
	)
	require.NoError(t, err)
	require.Equal(t, expectedNft, nftFromDB)
}

func TestNftHistoryChangedAfterNftMint(t *testing.T) {
	expectedNft := types.NftTransferQuery{
		ID:       nftID,
		DenomID:  burnableNftDenomID,
		NewOwner: User1,
		UniqID:   uniqBurnableNftId,
	}

	var nftFromDB types.NftTransferQuery
	err := config.QueryDatabase(`SELECT new_owner, old_owner FROM nft_transfer_history WHERE uniq_id = $1`, uniqBurnableNftId).Scan(
		&nftFromDB.NewOwner,
		&nftFromDB.OldOwner,
	)
	require.NoError(t, err)
	require.Equal(t, NoOwnerString, nftFromDB.OldOwner)
	require.Equal(t, expectedNft.NewOwner, nftFromDB.NewOwner)
}

func TestEditNft(t *testing.T) {
	// PREPARE
	oldUri := URI
	var oldUriFromDB string
	err := config.QueryDatabase(`SELECT URI FROM nft_nft WHERE id = $1 AND denom_id = $2`, nftID, burnableNftDenomID).Scan(&oldUriFromDB)
	require.NoError(t, err)
	require.Equal(t, oldUri, oldUriFromDB)

	newUri := "new URI"
	args := []string{
		nftModule,
		"edit",
		burnableNftDenomID,
		fmt.Sprintf("%d", nftID),
		config.GetFlag(URI, newUri),
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(User1, args...)
	require.NoError(t, err)

	// ASSERT
	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	var newUriFromDB string
	err = config.QueryDatabase(`SELECT URI FROM nft_nft WHERE id = $1 AND denom_id = $2`, nftID, burnableNftDenomID).Scan(&newUriFromDB)
	require.NoError(t, err)
	require.Equal(t, newUri, newUriFromDB)
}

func TestTransferNft(t *testing.T) {
	// PREPARE
	expectedCurrentOwner := User1
	intentedNewOwner := User3
	var currentOwnerFromDB string
	err := config.QueryDatabase(`SELECT owner FROM nft_nft WHERE id = $1 AND denom_id =$2`, nftID, burnableNftDenomID).Scan(&currentOwnerFromDB)
	require.NoError(t, err)
	require.Equal(t, expectedCurrentOwner, currentOwnerFromDB)
	transferArgs := []string{
		nftModule,
		"transfer",
		currentOwnerFromDB,
		intentedNewOwner,
		burnableNftDenomID,
		fmt.Sprintf("%d", nftID),
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(User1, transferArgs...)
	require.NoError(t, err)

	// ASSERT
	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	var newOwnerFromDB string
	err = config.QueryDatabase(`SELECT owner FROM nft_nft WHERE id = $1 AND denom_id =$2`, nftID, burnableNftDenomID).Scan(&newOwnerFromDB)
	require.NoError(t, err)
	require.Equal(t, intentedNewOwner, newOwnerFromDB)
}

func TestNftHistoryChangedAfterNftTransfer(t *testing.T) {
	expectedNewOwner := User3
	expectedOldOwner := User1
	var nftFromDB types.NftTransferQuery
	err := config.QueryDatabase(`SELECT new_owner, old_owner FROM nft_transfer_history WHERE uniq_id = $1 ORDER BY timestamp DESC`, uniqBurnableNftId).Scan(
		&nftFromDB.NewOwner,
		&nftFromDB.OldOwner,
	)
	require.NoError(t, err)
	require.Equal(t, expectedNewOwner, nftFromDB.NewOwner)
	require.Equal(t, expectedOldOwner, nftFromDB.OldOwner)
}

func TestBurnNft(t *testing.T) {
	// PREPARE
	burnArgs := []string{
		nftModule,
		"burn",
		burnableNftDenomID,
		fmt.Sprintf("%d", nftID),
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(User3, burnArgs...)
	require.NoError(t, err)

	// ASSERT BURN
	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	var burned bool
	err = config.QueryDatabase(`SELECT burned FROM nft_nft WHERE denom_id = $1 AND id = $2`, burnableNftDenomID, nftID).Scan(&burned)
	require.NoError(t, err)
	require.True(t, burned)
}

func TestNftHistoryChangedAfterBurn(t *testing.T) {
	expectedNewOwner := NoOwnerString
	expectedOldOwner := User3
	var nftFromDB types.NftTransferQuery
	err := config.QueryDatabase(`SELECT new_owner, old_owner FROM nft_transfer_history WHERE uniq_id = $1 ORDER BY timestamp DESC`, uniqBurnableNftId).Scan(
		&nftFromDB.NewOwner,
		&nftFromDB.OldOwner,
	)
	require.NoError(t, err)
	require.Equal(t, expectedNewOwner, nftFromDB.NewOwner)
	require.Equal(t, expectedOldOwner, nftFromDB.OldOwner)
}

func TestTransferDenom(t *testing.T) {
	//PREPARE
	intentedNewOwner := User3
	expectedCurrentOwner := User1
	var currentOwnerFromDB string
	err := config.QueryDatabase(`SELECT owner FROM nft_denom WHERE id = $1`, burnableNftDenomID).Scan(&currentOwnerFromDB)
	require.NoError(t, err)
	require.Equal(t, expectedCurrentOwner, currentOwnerFromDB)

	args := []string{
		nftModule,
		"transfer-denom",
		intentedNewOwner,
		burnableNftDenomID,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(expectedCurrentOwner, args...)
	require.NoError(t, err)

	// ASSERT
	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	var newOwnerFromDB string
	err = config.QueryDatabase(`SELECT owner FROM nft_denom WHERE id = $1`, burnableNftDenomID).Scan(&newOwnerFromDB)
	require.NoError(t, err)
	require.Equal(t, intentedNewOwner, newOwnerFromDB)
}
