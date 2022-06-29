package types

import cudoMintTypes "github.com/CudoVentures/cudos-node/x/cudoMint/types"

// MintParams represents the x/mint parameters
type MintParams struct {
	cudoMintTypes.GenesisState
	Height int64
}

// NewMintParams allows to build a new MintParams instance
func NewMintParams(params cudoMintTypes.GenesisState, height int64) *MintParams {
	return &MintParams{
		GenesisState: params,
		Height:       height,
	}
}
