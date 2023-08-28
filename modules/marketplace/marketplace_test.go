package marketplace

import (
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp/params"
	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v4/client/cryptocompare"
	"github.com/forbole/bdjuno/v4/database"
	mockutils "github.com/forbole/bdjuno/v4/mockutils"
	utils "github.com/forbole/bdjuno/v4/utils"
	"github.com/go-co-op/gocron"
	"github.com/stretchr/testify/suite"
)

var (
	timestamp    = time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
	one          = uint64(1)
	creator      = "Creator"
	denomId      = "DenomID"
	tokenId      = one
	oneStr       = "1"
	two          = uint64(2)
	twoStr       = "2"
	collectionID = int64(1)
	sdkCoin      = sdk.NewCoin("cudos", sdk.NewInt(1000))
	sdkCoinTwo   = sdk.NewCoin("cudos", sdk.NewInt(2000))
	dec, _       = sdkmath.LegacyNewDecFromStr(oneStr)
	royalty      = marketplaceTypes.Royalty{
		Address: oneStr,
		Percent: dec,
	}
	royalties          = []marketplaceTypes.Royalty{royalty}
	stringRoyalties, _ = json.Marshal(royalties)
)

type MarketplaceModuleTestSuite struct {
	suite.Suite
	module *Module
	db     *database.Db
}

func TestMarketplaceModuleTestSuite(t *testing.T) {
	suite.Run(t, new(MarketplaceModuleTestSuite))
}

func (suite *MarketplaceModuleTestSuite) SetupTest() {
	db, err := utils.NewTestDb("marketplaceTest")
	suite.Require().NoError(err)
	configBytes := []byte("testConfig")
	var cryptoCompareConfig cryptocompare.Config
	cryptoCompareClient := cryptocompare.NewClient(&cryptoCompareConfig)
	suite.module = NewModule(params.MakeTestEncodingConfig().Codec, db, configBytes, cryptoCompareClient)
	suite.module.cudosPrice.BTC = oneStr
	suite.module.cudosPrice.USD = twoStr
	suite.db = db

	_, err = db.SQL.Exec(`
INSERT INTO nft_denom (transaction_hash, id, name, schema, symbol, owner, contract_address_signer) 
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		oneStr, denomId, oneStr, oneStr, oneStr, oneStr, oneStr)
	suite.Require().NoError(err)
}

func (suite *MarketplaceModuleTestSuite) TestNewModule() {
	suite.Require().NotNil(suite.module)
	suite.Require().Equal("marketplace", suite.module.Name())
}

func (suite *MarketplaceModuleTestSuite) TestRegisterPeriodicOperations() {
	scheduler := gocron.NewScheduler(time.UTC)
	err := suite.module.RegisterPeriodicOperations(scheduler)
	suite.Require().NoError(err)
}

func (suite *MarketplaceModuleTestSuite) TestFetchCudosPrice() {
	err := suite.module.fetchCudosPrice()
	suite.Require().NoError(err)
}

func (suite *MarketplaceModuleTestSuite) TestHandleMsgs() {
	msgCreate := &marketplaceTypes.MsgCreateCollection{
		Creator:         creator,
		Id:              denomId,
		Name:            oneStr,
		Schema:          oneStr,
		Symbol:          oneStr,
		Traits:          oneStr,
		Description:     oneStr,
		Minter:          creator,
		Data:            oneStr,
		MintRoyalties:   royalties,
		ResaleRoyalties: royalties,
		Verified:        false,
	}
	msgPublish := &marketplaceTypes.MsgPublishCollection{
		Creator:         creator,
		DenomId:         denomId,
		MintRoyalties:   royalties,
		ResaleRoyalties: royalties,
	}
	msgVerify := &marketplaceTypes.MsgVerifyCollection{
		Creator: creator,
		Id:      one,
	}
	msgUnverify := &marketplaceTypes.MsgUnverifyCollection{
		Creator: creator,
		Id:      one,
	}
	msgUpdatePrice := &marketplaceTypes.MsgUpdatePrice{
		Creator: creator,
		Id:      one,
		Price:   sdkCoinTwo,
	}
	msgUpdateRoyalties := &marketplaceTypes.MsgUpdateRoyalties{
		Creator:         creator,
		Id:              one,
		MintRoyalties:   royalties,
		ResaleRoyalties: royalties,
	}
	msgMint := &marketplaceTypes.MsgMintNft{
		Creator:   creator,
		DenomId:   denomId,
		Recipient: oneStr,
		Price:     sdkCoin,
		Name:      oneStr,
		Uri:       oneStr,
		Data:      oneStr,
		Uid:       oneStr,
	}
	msgPublishNft := &marketplaceTypes.MsgPublishNft{
		Creator: creator,
		DenomId: denomId,
		TokenId: oneStr,
		Price:   sdkCoin,
	}
	msgBuy := &marketplaceTypes.MsgBuyNft{
		Creator: creator,
		Id:      one,
	}

	msgRemove := &marketplaceTypes.MsgRemoveNft{
		Creator: creator,
		Id:      one,
	}

	mockTX := mockutils.NewMockTxBuilder(suite.T(), timestamp, oneStr, one).
		WithEventPublishCollection(collectionID).
		WithEventCreateCollection(collectionID).
		WithEventVerifyCollection(collectionID).
		WithEventMintNft(tokenId).
		WithEventPublishNft(one).
		WithEventBuyNft(oneStr, denomId, creator).
		Build()

	err := suite.module.HandleMsg(0, msgCreate, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgPublish, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgVerify, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgUnverify, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgUpdatePrice, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgUpdateRoyalties, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgMint, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgPublishNft, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgBuy, mockTX)
	suite.Require().NoError(err)
	err = suite.module.HandleMsg(0, msgRemove, mockTX)
	suite.Require().NoError(err)
}
