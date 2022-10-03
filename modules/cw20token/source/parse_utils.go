package source

import (
	"encoding/json"
	"strconv"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/forbole/bdjuno/v2/types"
)

func ParseToTokenInfo(contractAddress string, res *wasmtypes.QueryAllContractStateResponse) (*types.TokenInfo, error) {
	// todo SAVE MARKETING INFO
	tokenInfo := types.TokenInfo{}
	for _, m := range res.Models {
		key := string(m.Key)

		if key == "token_info" {
			if err := json.Unmarshal(m.Value, &tokenInfo); err != nil {
				return nil, err
			}

			continue
		}

		if strings.Contains(key, "balance") {
			balance, err := strconv.ParseUint(strings.ReplaceAll(string(m.Value), "\"", ""), 10, 64)
			if err != nil {
				return nil, err
			}

			addressIndex := strings.Index(key, "cudos")
			address := key[addressIndex:]

			tokenInfo.Balances = append(tokenInfo.Balances, types.TokenBalance{Address: address, Amount: balance})
		}
	}

	tokenInfo.Address = contractAddress
	return &tokenInfo, nil
}

func ParseToBalance(res *wasmtypes.QuerySmartContractStateResponse) (uint64, error) {
	balance := struct {
		Balance uint64 `json:"balance,string"`
	}{}

	if err := json.Unmarshal(res.Data, &balance); err != nil {
		return 0, err
	}

	return balance.Balance, nil
}

func ParseToTotalSupply(res *wasmtypes.QuerySmartContractStateResponse) (uint64, error) {
	totalSupply := struct {
		TotalSupply uint64 `json:"total_supply,string"`
	}{}

	if err := json.Unmarshal(res.Data, &totalSupply); err != nil {
		return 0, err
	}

	return totalSupply.TotalSupply, nil
}
