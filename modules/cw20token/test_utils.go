package cw20token

import "github.com/forbole/bdjuno/v2/types"

const (
	nonTokenCodeID        = 1
	existingCodeID        = 2
	matchingCodeID        = 3
	nonTokenContract      = "1"
	existingTokenContract = "2"
	migratedTokenContract = "3"
	matchingTokenContract = "4"
)

func newPubMsg(codeID uint64, executeSchema string, querySchema string) *types.VerifiedContractPublishMessage {
	return &types.VerifiedContractPublishMessage{
		CodeID:        codeID,
		ExecuteSchema: executeSchema,
		QuerySchema:   querySchema,
	}
}

type sourceMock struct {
	tokenInfo *types.TokenInfo
	balance   uint64
	supply    uint64
}

func (s *sourceMock) GetTokenInfo(contract string, height int64) (*types.TokenInfo, error) {
	return s.tokenInfo, nil
}

func (s *sourceMock) GetBalance(contract string, address string, height int64) (uint64, error) {
	return s.balance, nil
}

func (s *sourceMock) GetCirculatingSupply(contract string, height int64) (uint64, error) {
	return s.supply, nil
}
