package cudomint

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"

	cudoMintTypes "github.com/CudoVentures/cudos-node/x/cudoMint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "cudomint").Msg("setting up periodic tasks")

	// // Setup a cron job to run every midnight
	if _, err := scheduler.Every(1).Minute().Do(func() {
		// if _, err := scheduler.Every(1).Day().At("00:00").Do(func() {
		utils.WatchMethod(m.calculateInflation)
		utils.WatchMethod(m.calculateAPR)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) calculateAPR() error {
	mintParams, err := m.db.GetMintParams()
	if err != nil {
		return err
	}

	minter := mintParams.Minter

	if minter.NormTimePassed.GT(finalNormTimePassed) {
		return nil
	}

	lastBlockHeight, err := m.db.GetLastBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get last block height %s", err)
	}

	mintAmountInt, err := m.calculateMintedTokensSinceHeight(lastBlockHeight, 30*12)
	if err != nil {
		return fmt.Errorf("failed to calculated minted tokens: %s", err)
	}

	bondedTokens, err := m.db.GetBondedTokens()
	if err != nil {
		return fmt.Errorf("failed to get bonded_tokens: %s", err)
	}

	apr := mintAmountInt.ToDec().Quo(bondedTokens.ToDec())

	if err := m.db.SaveAPR(apr, lastBlockHeight); err != nil {
		return fmt.Errorf("failed to save apr: %s", err)
	}

	if err := m.db.SaveAPRHistory(apr, lastBlockHeight, time.Now().UnixNano()); err != nil {
		return fmt.Errorf("failed to save apr history: %s", err)
	}

	return nil
}

func (m *Module) calculateInflation() error {
	client, err := ethclient.Dial(m.config.EthNode)
	if err != nil {
		return fmt.Errorf("failed to dial eth node: %s", err)
	}

	latestEthBlock, err := getLatestEthBlock(client)
	if err != nil {
		return fmt.Errorf("faield to get latest eth block: %s", err)
	}

	currentTotalBalance, err := getEthAccountsBalanceAtBlock(client, m.config.TokenAddress, m.config.EthAccounts, latestEthBlock)
	if err != nil {
		return fmt.Errorf("failed to get eth accounts balance: %s", err)
	}

	inflationStartBlock := latestEthBlock
	inflationStartBlock.Sub(latestEthBlock, big.NewInt(inflationSinceDays*ethBlocksPerDay))

	startTotalBalance, err := getEthAccountsBalanceAtBlock(client, m.config.TokenAddress, m.config.EthAccounts, inflationStartBlock)
	if err != nil {
		return fmt.Errorf("failed to get eth accounts balance: %s", err)
	}

	lastBlockHeight, err := m.db.GetLastBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get last block height %s", err)
	}

	mintParams, err := m.db.GetMintParams()
	if err != nil {
		return err
	}

	startBlockHeight := int64(0)
	inflationSinceDays := int64(inflationSinceDays)

	// TODO: This can be removed after the chain is working for more than INFLATION_SINCE_DAYS

	for startBlockHeight < 1 {
		startBlockHeight = lastBlockHeight - (inflationSinceDays * mintParams.Params.BlocksPerDay.Int64())
		inflationSinceDays--
	}

	mintAmountInt, err := m.calculateMintedTokensSinceHeight(startBlockHeight, 30)
	if err != nil {
		return fmt.Errorf("failed to calculated minted tokens: %s", err)
	}

	startTotalBalanceInt, ok := sdk.NewIntFromString(startTotalBalance.String())
	if !ok {
		return fmt.Errorf("failed to convert big.Int to sdk.Int: %s", startTotalBalance.String())
	}

	startTotalSupply, _ := sdk.NewIntFromString(maxSupply)
	startTotalSupply = startTotalSupply.Sub(startTotalBalanceInt)

	currentTotalBalanceInt, ok := sdk.NewIntFromString(currentTotalBalance.String())
	if !ok {
		return fmt.Errorf("failed to convert big.Int to sdk.Int: %s", currentTotalBalance.String())
	}

	currentTotalBalanceInt = currentTotalBalanceInt.Add(mintAmountInt)

	currentTotalSupply, _ := sdk.NewIntFromString(maxSupply)
	currentTotalSupply = currentTotalSupply.Sub(currentTotalBalanceInt)

	inflation := currentTotalSupply.Sub(startTotalSupply).ToDec().Quo(startTotalSupply.ToDec())

	if err := m.db.SaveInflation(inflation, lastBlockHeight); err != nil {
		return fmt.Errorf("failed to store inflation: %s", err)
	}

	if err := m.db.SaveAdjustedSupply(currentTotalSupply.ToDec(), lastBlockHeight); err != nil {
		return fmt.Errorf("failed to store adjusted supply: %s", err)
	}

	return nil
}

func getLatestEthBlock(client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest eth block: %s", err)
	}

	return header.Number, nil
}

func getEthAccountsBalanceAtBlock(client *ethclient.Client, tokenAddress string, accounts []string, block *big.Int) (*big.Int, error) {
	instance, err := NewTokenCaller(common.HexToAddress(tokenAddress), client)
	if err != nil {
		return nil, err
	}

	totalBalance := big.NewInt(0)

	for _, account := range accounts {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		balance, err := instance.BalanceOf(&bind.CallOpts{
			BlockNumber: block,
			Context:     ctx,
		}, common.HexToAddress(account))

		if err != nil {
			return nil, err
		}

		totalBalance.Add(totalBalance, balance)
	}

	return totalBalance, nil
}

func (m *Module) calculateMintedTokensSinceHeight(sinceBlock, periodDays int64) (sdk.Int, error) {
	genesis, err := m.db.GetGenesis()
	if err != nil {
		return sdk.Int{}, fmt.Errorf("failed to get genesis: %s", err)
	}

	mintParams, err := m.db.GetMintParams()
	if err != nil {
		return sdk.Int{}, fmt.Errorf("failed to get mint params: %s", err)
	}

	minter := mintParams.Minter
	params := mintParams.Params

	if minter.NormTimePassed.GT(finalNormTimePassed) {
		return sdk.NewInt(0), nil
	}

	minter.NormTimePassed = updateNormTimePassed(mintParams, genesis.InitialHeight, sinceBlock)

	mintAmountInt := sdk.NewInt(0)
	totalBlocks := mintParams.Params.BlocksPerDay.Int64() * periodDays

	for height := int64(1); height <= totalBlocks; height++ {
		if minter.NormTimePassed.GT(finalNormTimePassed) {
			break
		}

		incr := normalizeBlockHeightInc(params.BlocksPerDay)
		mintAmountDec := calculateMintedCoins(minter, incr)
		mintAmountInt = mintAmountInt.Add(mintAmountDec.TruncateInt())
		minter.NormTimePassed.Add(incr)
	}

	return mintAmountInt, nil
}

func updateNormTimePassed(mintParams types.MintParams, initialBlockHeight, lastBlockHeight int64) sdk.Dec {
	// TODO: Cannot be saved at this moment because of the changes in inflation calculation
	// storage := workers.NewWorkersStorage(db, "cudomint")
	// valueStr, err := storage.GetOrDefaultValue(calculateInflationLastBlock, strconv.FormatInt(initialBlockHeight, 10))
	// if err != nil {
	// 	return sdk.Dec{}, fmt.Errorf("failed to get %s", calculateInflationLastBlock)
	// }

	// value, err := strconv.ParseInt(valueStr, 10, 64)
	// if err != nil {
	// 	return sdk.Dec{}, fmt.Errorf("failed to parse %s", calculateInflationLastBlock)
	// }

	for initialBlockHeight < lastBlockHeight {
		inc := normalizeBlockHeightInc(mintParams.Params.BlocksPerDay)
		mintParams.Minter.NormTimePassed.Add(inc)
		initialBlockHeight++
	}

	// if err := db.SaveMintParams(&mintParams); err != nil {
	// 	return sdk.Dec{}, fmt.Errorf("failed to save mint params: %s", err)
	// }

	// if err := storage.SetValue(calculateInflationLastBlock, strconv.FormatInt(lastBlockHeight, 10)); err != nil {
	// 	return sdk.Dec{}, fmt.Errorf("failed to save %s: %s", calculateInflationLastBlock, err)
	// }

	return mintParams.Minter.NormTimePassed
}

// Normalize block height incrementation
func normalizeBlockHeightInc(incrementModifier sdk.Int) sdk.Dec {
	totalBlocks := incrementModifier.Mul(totalDays)
	return (sdk.NewDec(1).QuoInt(totalBlocks)).Mul(finalNormTimePassed)
}

// Integral of f(t) is 0,6 * t^3  - 26.5 * t^2 + 358 * t
// The function extrema is ~10.48 so after that the function is decreasing
func calculateIntegral(t sdk.Dec) sdk.Dec {
	return (zeroPointSix.Mul(t.Power(3))).Sub(twentySixPointFive.Mul(t.Power(2))).Add(sdk.NewDec(358).Mul(t))
}

func calculateMintedCoins(minter cudoMintTypes.Minter, increment sdk.Dec) sdk.Dec {
	prevStep := calculateIntegral(sdk.MinDec(minter.NormTimePassed, finalNormTimePassed))
	nextStep := calculateIntegral(sdk.MinDec(minter.NormTimePassed.Add(increment), finalNormTimePassed))
	return (nextStep.Sub(prevStep)).Mul(sdk.NewDec(10).Power(24)) // formula calculates in mil of cudos + converting to acudos
}

var (
	// based on the assumption that we have 1 block per 5 seconds
	// if actual blocks are generated at slower rate then the network will mint tokens more than 3652 days (~10 years)
	totalDays           = sdk.NewInt(3652) // Hardcoded to 10 years
	finalNormTimePassed = sdk.NewDec(10)
	zeroPointSix        = sdk.MustNewDecFromStr("0.6")
	twentySixPointFive  = sdk.MustNewDecFromStr("26.5")
	// calculateInflationLastBlock = "CalculateInflationLastBlock"
)

const ethBlocksPerDay = 5760
const inflationSinceDays = 30 * 3
const maxSupply = "10000000000000000000000000000" // 10 billion
