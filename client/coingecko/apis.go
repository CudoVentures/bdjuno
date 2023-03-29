package coingecko

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"

	"github.com/forbole/bdjuno/v2/types"
)

// GetTokensPrices queries the remote APIs to get the token prices of all the tokens having the given ids
func GetTokensPrices(currency string, ids []string) ([]types.TokenPrice, error) {
	var prices []MarketTicker
	query := fmt.Sprintf("/data/pricemultifull?fsyms=%stsyms=%s", currency, strings.Join(ids, ","))
	err := queryCoinGecko(query, &prices)
	if err != nil {
		return nil, err
	}

	return ConvertCoingeckoPrices(prices), nil
}

func ConvertCoingeckoPrices(prices []MarketTicker) []types.TokenPrice {
	tokenPrices := make([]types.TokenPrice, len(prices))
	for i, price := range prices {
		tokenPrices[i] = types.NewTokenPrice(
			price.Symbol,
			price.CurrentPrice,
			int64(math.Trunc(price.MarketCap)),
			price.LastUpdated,
		)
	}
	return tokenPrices
}

func GetCUDOSPrice(currency string) (string, error) {
	ids := []string{"CUDOS"}
	prices, err := GetTokensPrices(currency, ids)
	if err != nil {
		return "", err
	}
	price := fmt.Sprintf("%g", prices[0].Price)
	return price, err
}

// queryCoinGecko queries the CoinGecko APIs for the given endpoint
func queryCoinGecko(endpoint string, ptr interface{}) error {
	req, err := http.NewRequest("GET", "https://min-api.cryptocompare.com"+endpoint, nil)

	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
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
