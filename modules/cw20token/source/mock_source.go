package source

import (
	"encoding/json"
	"fmt"
	"strconv"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/bdjuno/v2/types"
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

func (s *MockSource) GetTokenInfo(token string, height int64) (*wasm.QueryAllContractStateResponse, error) {
	res := &wasm.QueryAllContractStateResponse{}

	tokenInfo := s.T
	tokenInfo.Balances = []types.TokenBalance{}
	tokenInfo.MarketingInfo = types.MarketingInfo{}

	tokenInfoJson, err := json.Marshal(tokenInfo)
	if err != nil {
		return nil, err
	}

	res.Models = append(res.Models, wasm.Model{
		Key:   []byte("token_info"),
		Value: tokenInfoJson,
	})

	marketingInfo, err := json.Marshal(s.T.MarketingInfo)
	if err != nil {
		return nil, err
	}

	res.Models = append(res.Models, wasm.Model{
		Key:   []byte("marketing_info"),
		Value: marketingInfo,
	})

	res.Models = append(res.Models, wasm.Model{
		Key:   []byte("logo"),
		Value: []byte(s.T.Logo),
	})

	for _, b := range s.T.Balances {
		res.Models = append(res.Models, wasm.Model{
			Key:   []byte(fmt.Sprintf("balance%s", b.Address)),
			Value: []byte(strconv.FormatUint(b.Amount, 10)),
		})
	}

	return res, nil
}

func (s *MockSource) GetBalance(token string, address string, height int64) (*wasm.QuerySmartContractStateResponse, error) {
	var balance uint64
	for _, b := range s.T.Balances {
		if b.Address == address {
			balance = b.Amount
		}
	}

	return &wasm.QuerySmartContractStateResponse{
		Data: []byte(fmt.Sprintf(`{"balance":"%d"}`, balance)),
	}, nil
}

func (s *MockSource) GetCirculatingSupply(token string, height int64) (*wasm.QuerySmartContractStateResponse, error) {
	data := []byte(fmt.Sprintf(`{"total_supply":"%d"}`, s.T.CirculatingSupply))
	return &wasm.QuerySmartContractStateResponse{Data: data}, nil
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

func (s *MockSource) Transfer(sender string, recipient string, amount uint64) error {
	i := s.getBalanceIndex(sender)
	if s.T.Balances[i].Amount < amount {
		return fmt.Errorf("insufficient balance")
	}

	s.T.Balances[i].Amount -= amount

	if s.T.Balances[i].Amount == 0 {
		s.T.Balances = append(s.T.Balances[:i], s.T.Balances[i+1:]...)
	}

	i = s.getBalanceIndex(recipient)
	s.T.Balances[i].Amount += amount

	return nil
}

func (s *MockSource) Burn(sender string, amount uint64) error {
	i := s.getBalanceIndex(sender)
	if s.T.Balances[i].Amount < amount {
		return fmt.Errorf("insufficient balance")
	}

	s.T.Balances[i].Amount -= amount

	if s.T.Balances[i].Amount == 0 {
		s.T.Balances = append(s.T.Balances[:i], s.T.Balances[i+1:]...)
	}

	s.T.CirculatingSupply -= amount

	return nil
}

func (s *MockSource) Mint(recipient string, amount uint64) error {
	if s.T.CirculatingSupply+amount > s.T.MintInfo.MaxSupply {
		return fmt.Errorf("cannot exceed max supply")
	}

	i := s.getBalanceIndex(recipient)
	s.T.Balances[i].Amount += amount
	s.T.CirculatingSupply += amount

	return nil
}

func (s *MockSource) UpdateMinter(newMinter string) {
	s.T.MintInfo.Minter = newMinter
}

func (s *MockSource) UpdateLogo(newLogo string) {
	s.T.Logo = newLogo
}

func (s *MockSource) UpdateMarketingInfo(newMarketingInfo types.MarketingInfo) {
	s.T.MarketingInfo = newMarketingInfo
}
