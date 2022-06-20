package cudomint

import (
	"fmt"
	"strconv"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"

	cudoMintTypes "github.com/CudoVentures/cudos-node/x/cudoMint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	workers "github.com/forbole/bdjuno/v2/workers"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "cudomint").Msg("setting up periodic tasks")

	// // Setup a cron job to run every midnight
	if _, err := scheduler.Every(1).Day().At("00:00").Do(func() {
		utils.WatchMethod(m.calculateInflation)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) calculateInflation() error {
	genesis, err := m.db.GetGenesis()
	if err != nil {
		return fmt.Errorf("failed to get genesis: %s", err)
	}

	mintParams, err := m.db.GetMintParams()
	if err != nil {
		return err
	}

	minter := mintParams.Minter
	params := mintParams.Params

	if minter.NormTimePassed.GT(FinalNormTimePassed) {
		return nil
	}

	lastBlockHeight, err := m.db.GetLastBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get last block height %s", err)
	}

	supply, err := m.source.GetSupply(lastBlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get supply: %s", err)
	}

	minter.NormTimePassed, err = updateNormTimePassed(m.db, mintParams, genesis.InitialHeight, lastBlockHeight)
	if err != nil {
		return fmt.Errorf("failed to update normTimePassed: %s", err)
	}

	mintAmountInt := sdk.NewInt(0)
	totalBlocks := mintParams.Params.BlocksPerDay.Int64() * 30 * 12 // 1 year

	for height := int64(1); height <= totalBlocks; height++ {
		if minter.NormTimePassed.GT(FinalNormTimePassed) {
			break
		}

		incr := normalizeBlockHeightInc(params.BlocksPerDay)
		mintAmountDec := calculateMintedCoins(minter, incr)
		mintAmountInt = mintAmountInt.Add(mintAmountDec.TruncateInt())
		minter.NormTimePassed.Add(incr)
	}

	inflation := mintAmountInt.ToDec().Quo(supply.AmountOf("acudos").ToDec())

	if err := m.db.SaveInflation(inflation, lastBlockHeight); err != nil {
		return fmt.Errorf("failed to save inflation: %s", err)
	}

	return nil
}

func updateNormTimePassed(db *database.Db, mintParams types.MintParams, initialBlockHeight, lastBlockHeight int64) (sdk.Dec, error) {
	storage := workers.NewWorkersStorage(db, "cudomint")
	valueStr, err := storage.GetOrDefaultValue(calculateInflationLastBlock, strconv.FormatInt(initialBlockHeight, 10))
	if err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to get %s", calculateInflationLastBlock)
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to parse %s", calculateInflationLastBlock)
	}

	for value < lastBlockHeight {
		inc := normalizeBlockHeightInc(mintParams.Params.BlocksPerDay)
		mintParams.Minter.NormTimePassed.Add(inc)
		value++
	}

	if err := db.SaveMintParams(&mintParams); err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to save mint params: %s", err)
	}

	if err := storage.SetValue(calculateInflationLastBlock, strconv.FormatInt(lastBlockHeight, 10)); err != nil {
		return sdk.Dec{}, fmt.Errorf("failed to save %s: %s", calculateInflationLastBlock, err)
	}

	return mintParams.Minter.NormTimePassed, nil
}

// Normalize block height incrementation
func normalizeBlockHeightInc(incrementModifier sdk.Int) sdk.Dec {
	totalBlocks := incrementModifier.Mul(totalDays)
	return (sdk.NewDec(1).QuoInt(totalBlocks)).Mul(FinalNormTimePassed)
}

// Integral of f(t) is 0,6 * t^3  - 26.5 * t^2 + 358 * t
// The function extrema is ~10.48 so after that the function is decreasing
func calculateIntegral(t sdk.Dec) sdk.Dec {
	return (zeroPointSix.Mul(t.Power(3))).Sub(twentySixPointFive.Mul(t.Power(2))).Add(sdk.NewDec(358).Mul(t))
}

func calculateMintedCoins(minter cudoMintTypes.Minter, increment sdk.Dec) sdk.Dec {
	prevStep := calculateIntegral(sdk.MinDec(minter.NormTimePassed, FinalNormTimePassed))
	nextStep := calculateIntegral(sdk.MinDec(minter.NormTimePassed.Add(increment), FinalNormTimePassed))
	return (nextStep.Sub(prevStep)).Mul(sdk.NewDec(10).Power(24)) // formula calculates in mil of cudos + converting to acudos
}

var (
	// based on the assumption that we have 1 block per 5 seconds
	// if actual blocks are generated at slower rate then the network will mint tokens more than 3652 days (~10 years)
	totalDays                   = sdk.NewInt(3652) // Hardcoded to 10 years
	FinalNormTimePassed         = sdk.NewDec(10)
	zeroPointSix                = sdk.MustNewDecFromStr("0.6")
	twentySixPointFive          = sdk.MustNewDecFromStr("26.5")
	calculateInflationLastBlock = "CalculateInflationLastBlock"
)
