package cryptoCompare

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/forbole/bdjuno/v2/types"
)

// GetTokensPrices queries the remote APIs to get the token prices of all the tokens having the given ids
func GetTokensPrices(currency string, ids []string, apiKey string) ([]types.TokenPrice, error) {
	var resStruct struct {
		Raw map[string]map[string]MarketTicker
	}
	query := fmt.Sprintf("/data/pricemultifull?fsyms=%s&tsyms=%s", strings.Join(ids, ","), currency)
	err := queryCoinGecko(query, &resStruct, apiKey)
	if err != nil {
		return nil, err
	}

	// return nil, nil
	return ConvertCoingeckoPrices(resStruct.Raw), nil
}

func ConvertCoingeckoPrices(tokens map[string]map[string]MarketTicker) []types.TokenPrice {
	var tokenPrices []types.TokenPrice

	for token, price := range tokens {
		for _, marketTicker := range price {
			tokenPrices = append(tokenPrices, types.NewTokenPrice(
				strings.ToLower(token),
				marketTicker.CurrentPrice,
				int64(math.Trunc(marketTicker.MarketCap)),
				time.Unix(marketTicker.LastUpdated, 0),
			))
		}
	}
	return tokenPrices
}
func GetCUDOSPrice(currency string, apiKey string) (string, error) {
	ids := []string{"CUDOS"}
	prices, err := GetTokensPrices(currency, ids, apiKey)
	if err != nil {
		return "", err
	}
	price := fmt.Sprintf("%g", prices[0].Price)
	return price, err
}

// queryCoinGecko queries the CoinGecko APIs for the given endpoint
func queryCoinGecko(endpoint string, ptr interface{}, apiKey string) error {
	req, err := http.NewRequest("GET", "https://min-api.cryptocompare.com"+endpoint, nil)

	if err != nil {
		return err
	}

	if apiKey != "" {
		log.Debug().Str("module", "crypto-compare").Msg("using api key")
		req.Header.Set("authorization", fmt.Sprintf("Apikey %s", apiKey))
	} else {
		log.Debug().Str("module", "crypto-compare").Msg("no api key provided")
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response body: %s", err)
	}

	err = json.Unmarshal(bz, &ptr)
	if err != nil {
		return fmt.Errorf("error while unmarshaling response body: %s", err)
	}

	return nil
}
