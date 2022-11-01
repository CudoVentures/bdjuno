package source

import (
	"encoding/json"

	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
)

var (
	_ source.Source = &MockSource{}
)

type MockSource struct {
	T types.TokenInfo
}

func NewMockSource(token types.TokenInfo) *MockSource {
	tokenCopy := token
	tokenCopy.Balances = []types.TokenBalance{}
	tokenCopy.Balances = append(tokenCopy.Balances, token.Balances...)

	return &MockSource{tokenCopy}
}

func (s *MockSource) TokenInfo(tokenAddr string, height int64) (types.TokenInfo, error) {
	return s.T, nil
}

func (s *MockSource) AllBalances(tokenAddr string, height int64) ([]types.TokenBalance, error) {
	return s.T.Balances, nil
}

func (s *MockSource) Balance(tokenAddr string, address string, height int64) (uint64, error) {
	for _, b := range s.T.Balances {
		if b.Address == address {
			return b.Amount, nil
		}
	}

	return 0, nil
}

func (s *MockSource) TotalSupply(tokenAddr string, height int64) (uint64, error) {
	return s.T.TotalSupply, nil
}

func (s *MockSource) getBalanceIndex(addr string) int {
	for i, b := range s.T.Balances {
		if b.Address == addr {
			return i
		}
	}

	s.T.Balances = append(s.T.Balances, types.TokenBalance{Address: addr})
	return len(s.T.Balances) - 1
}

func (s *MockSource) Transfer(sender string, recipient string, amount uint64) {
	i := s.getBalanceIndex(sender)

	s.T.Balances[i].Amount -= amount

	if s.T.Balances[i].Amount == 0 {
		s.T.Balances = append(s.T.Balances[:i], s.T.Balances[i+1:]...)
	}

	i = s.getBalanceIndex(recipient)
	s.T.Balances[i].Amount += amount
}

func (s *MockSource) Burn(sender string, amount uint64) {
	i := s.getBalanceIndex(sender)

	s.T.Balances[i].Amount -= amount

	if s.T.Balances[i].Amount == 0 {
		s.T.Balances = append(s.T.Balances[:i], s.T.Balances[i+1:]...)
	}

	s.T.TotalSupply -= amount
}

func (s *MockSource) Mint(recipient string, amount uint64) {
	i := s.getBalanceIndex(recipient)
	s.T.Balances[i].Amount += amount
	s.T.TotalSupply += amount
}

func (s *MockSource) UpdateMinter(newMinter string) {
	s.T.Mint.Minter = newMinter
}

func (s *MockSource) UpdateLogo(newLogo string) {
	s.T.Marketing.Logo = json.RawMessage(newLogo)
}

func (s *MockSource) UpdateMarketing(marketing types.Marketing) {
	s.T.Marketing = marketing
}
