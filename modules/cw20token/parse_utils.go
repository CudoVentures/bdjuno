package cw20token

import (
	"encoding/json"
	"strconv"
	"strings"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
)

func parseToTokenInfo(res *wasm.QueryAllContractStateResponse) (*types.TokenInfo, error) {
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

			continue
		}

		if key == "marketing_info" {
			if err := json.Unmarshal(m.Value, &tokenInfo.MarketingInfo); err != nil {
				return nil, err
			}
			continue
		}

		if key == "logo" {
			tokenInfo.Logo = utils.SanitizeUTF8(string(m.Value))
			continue
		}
	}

	return &tokenInfo, nil
}

func parseToBalance(res *wasm.QuerySmartContractStateResponse) (uint64, error) {
	balance := struct {
		Balance uint64 `json:"balance,string"`
	}{}

	if err := json.Unmarshal(res.Data, &balance); err != nil {
		return 0, err
	}

	return balance.Balance, nil
}

func parseToCirculatingSupply(res *wasm.QuerySmartContractStateResponse) (uint64, error) {
	totalSupply := struct {
		TotalSupply uint64 `json:"total_supply,string"`
	}{}

	if err := json.Unmarshal(res.Data, &totalSupply); err != nil {
		return 0, err
	}

	return totalSupply.TotalSupply, nil
}

func parseToMsgExecuteToken(msg *wasm.MsgExecuteContract) (*types.MsgExecuteToken, error) {
	req := map[string]json.RawMessage{}
	if err := json.Unmarshal(msg.Msg, &req); err != nil {
		return nil, err
	}

	res := types.MsgExecuteToken{}
	for key, val := range req {
		if err := json.Unmarshal(val, &res); err != nil {
			return nil, err
		}

		res.Type = key
		res.MsgRaw = val
	}

	res.Contract = msg.Contract
	res.Sender = msg.Sender

	return &res, nil
}
