package source

import (
	"encoding/json"
	"strconv"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
)

var (
	_ source.Source = &MockSource{}
)

type MockSource struct {
	T types.TokenInfo
	A []dbtypes.AllowanceRow
}

func NewMockSource(token types.TokenInfo, allowances []dbtypes.AllowanceRow) *MockSource {
	tokenCopy := token
	tokenCopy.Balances = []types.TokenBalance{}
	tokenCopy.Balances = append(tokenCopy.Balances, token.Balances...)

	allowancesCopy := []dbtypes.AllowanceRow{}
	allowancesCopy = append(allowancesCopy, allowances...)

	return &MockSource{tokenCopy, allowancesCopy}
}

func (s *MockSource) TokenInfo(tokenAddr string, height int64) (types.TokenInfo, error) {
	return s.T, nil
}

func (s *MockSource) AllBalances(tokenAddr string, height int64) ([]types.TokenBalance, error) {
	return s.T.Balances, nil
}

func (s *MockSource) Balance(tokenAddr string, address string, height int64) (string, error) {
	for _, b := range s.T.Balances {
		if b.Address == address {
			return b.Amount, nil
		}
	}

	return "0", nil
}

func (s *MockSource) TotalSupply(tokenAddr string, height int64) (string, error) {
	return s.T.TotalSupply, nil
}

func (s *MockSource) Allowance(tokenAddr string, owner string, spender string, height int64) (types.Allowance, error) {
	for _, a := range s.A {
		if a.Owner == owner && a.Spender == spender {
			return types.Allowance{a.Amount, json.RawMessage(a.Expires)}, nil
		}
	}
	return types.Allowance{"0", []byte(nil)}, nil
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

	balance, _ := strconv.ParseUint(s.T.Balances[i].Amount, 10, 64)
	balance -= amount
	s.T.Balances[i].Amount = strconv.FormatUint(balance, 10)

	if s.T.Balances[i].Amount == "0" {
		s.T.Balances = append(s.T.Balances[:i], s.T.Balances[i+1:]...)
	}

	i = s.getBalanceIndex(recipient)
	balance, _ = strconv.ParseUint(s.T.Balances[i].Amount, 10, 64)
	balance += amount
	s.T.Balances[i].Amount = strconv.FormatUint(balance, 10)
}

func (s *MockSource) Burn(sender string, amount uint64) {
	i := s.getBalanceIndex(sender)

	balance, _ := strconv.ParseUint(s.T.Balances[i].Amount, 10, 64)
	balance -= amount
	s.T.Balances[i].Amount = strconv.FormatUint(balance, 10)

	if s.T.Balances[i].Amount == "0" {
		s.T.Balances = append(s.T.Balances[:i], s.T.Balances[i+1:]...)
	}

	totalSupply, _ := strconv.ParseUint(s.T.TotalSupply, 10, 64)
	totalSupply -= amount
	s.T.TotalSupply = strconv.FormatUint(totalSupply, 10)

}

func (s *MockSource) Mint(recipient string, amount uint64) {
	i := s.getBalanceIndex(recipient)

	balance, _ := strconv.ParseUint(s.T.Balances[i].Amount, 10, 64)
	balance += amount
	s.T.Balances[i].Amount = strconv.FormatUint(balance, 10)

	totalSupply, _ := strconv.ParseUint(s.T.TotalSupply, 10, 64)
	totalSupply += amount
	s.T.TotalSupply = strconv.FormatUint(totalSupply, 10)
}

func (s *MockSource) UpdateMinter(newMinter string) {
	s.T.Mint.Minter = newMinter
}

func (s *MockSource) UpdateLogo(newLogo string) {
	logo := json.RawMessage(newLogo)
	s.T.Marketing.Logo = &logo
}

func (s *MockSource) UpdateMarketing(marketing types.Marketing) {
	s.T.Marketing = marketing
}

func (s *MockSource) IncreaseAllowance(owner string, spender string, amount uint64, expires json.RawMessage) {
	for i, a := range s.A {
		if a.Owner == owner && a.Spender == spender {
			allowance, _ := strconv.ParseUint(a.Amount, 10, 64)
			allowance += amount
			s.A[i].Amount = strconv.FormatUint(allowance, 10)
			s.A[i].Expires = string(expires)
			return
		}
	}
	s.A = append(s.A, dbtypes.AllowanceRow{s.T.Address, owner, spender, strconv.FormatUint(amount, 10), string(expires)})
}

func (s *MockSource) DecreaseAllowance(owner string, spender string, amount uint64, expires json.RawMessage) {
	for i, a := range s.A {
		if a.Owner == owner && a.Spender == spender {
			allowance, _ := strconv.ParseUint(a.Amount, 10, 64)
			allowance -= amount
			s.A[i].Amount = strconv.FormatUint(allowance, 10)
			s.A[i].Expires = string(expires)
			if allowance <= 0 {
				s.A = append(s.A[:i], s.A[i+1:]...)
			}
		}
	}
}
