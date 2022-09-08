package parsegenesis

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	parsecmdtypes "github.com/forbole/juno/v3/cmd/parse/types"
	parsetypes "github.com/forbole/juno/v3/cmd/parse/types"
	"github.com/forbole/juno/v3/modules"
	"github.com/forbole/juno/v3/parser"
	"github.com/forbole/juno/v3/types"
	"github.com/forbole/juno/v3/types/config"
	junoutils "github.com/forbole/juno/v3/types/utils"
	"github.com/spf13/cobra"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// NewParseGenesisCmd returns the Cobra command allowing to parse the genesis file
func NewParseGenesisCmd(parseCfg *parsetypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "parse-genesis [[module names]]",
		Short:   "Parse genesis file. To parse specific modules, input module names as arguments",
		Example: "bdjuno parse-genesis auth bank consensus gov history staking",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return parsetypes.UpdatedGlobalCfg(parseCfg)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseCfg)
			if err != nil {
				return err
			}

			// Get the modules to parse
			var modulesToParse []modules.Module
			for _, moduleName := range args {
				module, found := getModule(moduleName, parseCtx)
				if !found {
					return fmt.Errorf("module %s is not registered", moduleName)
				}

				modulesToParse = append(modulesToParse, module)
			}

			// Default to all the modules
			if len(modulesToParse) == 0 {
				modulesToParse = parseCtx.Modules
			}

			// Get the genesis doc and state
			genesisDoc, genesisState, err := junoutils.GetGenesisDocAndState(config.Cfg.Parser.GenesisFilePath, parseCtx.Node)
			if err != nil {
				return fmt.Errorf("error while getting genesis doc and state: %s", err)
			}

			if err := resolveInitialHeightBlock(parseCtx, genesisDoc.InitialHeight); err != nil {
				return fmt.Errorf("failed to resolve initial height block: %v", err)
			}

			// For each module, parse the genesis
			for _, module := range modulesToParse {
				if genesisModule, ok := module.(modules.GenesisModule); ok {
					err = genesisModule.HandleGenesis(genesisDoc, genesisState)
					if err != nil {
						return fmt.Errorf("error while parsing genesis of %s module: %s", module.Name(), err)
					}
				}
			}

			return nil
		},
	}
}

// doesModuleExist tells whether a module with the given name exist inside the specified context ot not
func getModule(module string, parseCtx *parser.Context) (modules.Module, bool) {
	for _, mod := range parseCtx.Modules {
		if module == mod.Name() {
			return mod, true
		}
	}
	return nil, false
}

func resolveInitialHeightBlock(ctx *parser.Context, initialHeight int64) error {
	hasBlock, err := ctx.Database.HasBlock(initialHeight)
	if err != nil {
		return err
	}
	if hasBlock {
		return nil
	}

	block, err := ctx.Node.Block(initialHeight)
	if err != nil {
		return fmt.Errorf("failed to get block from node: %s", err)
	}

	if err := resolveInitialHeightValidator(ctx, initialHeight, block); err != nil {
		return err
	}

	txs, err := ctx.Node.Txs(block)
	if err != nil {
		return fmt.Errorf("failed to get transactions for block: %s", err)
	}

	var totalGas uint64
	for _, tx := range txs {
		totalGas += uint64(tx.GasUsed)
	}

	return ctx.Database.SaveBlock(types.NewBlockFromTmBlock(block, totalGas))
}

func resolveInitialHeightValidator(ctx *parser.Context, initialHeight int64, block *coretypes.ResultBlock) error {
	vals, err := ctx.Node.Validators(initialHeight)
	if err != nil {
		return fmt.Errorf("failed to get validators for block: %s", err)
	}

	isValidatorFound := false
	proposerAddr := sdk.ConsAddress(block.Block.ProposerAddress)
	for _, val := range vals.Validators {
		if proposerAddr.String() == sdk.ConsAddress(val.Address).String() {
			isValidatorFound = true

			consAddr := sdk.ConsAddress(val.Address).String()
			consPubKey, err := types.ConvertValidatorPubKeyToBech32String(val.PubKey)
			if err != nil {
				return fmt.Errorf("failed to convert validator public key for validators %s: %s", consAddr, err)
			}

			validators := make([]*types.Validator, 1)
			validators[0] = types.NewValidator(consAddr, consPubKey)

			if err := ctx.Database.SaveValidators(validators); err != nil {
				return err
			}
		}
	}

	if !isValidatorFound {
		return fmt.Errorf("validator %s not found", proposerAddr.String())
	}

	return nil
}
