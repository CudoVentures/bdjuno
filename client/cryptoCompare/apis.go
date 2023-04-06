package cryptoCompare

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/forbole/bdjuno/v2/types"
)

// NewModule returns a new Module instance
func NewClient(cfg *Config) *CryptoCompareClient {
	return &CryptoCompareClient{
		useProdApiKey: true,
		config:        cfg,
	}
}

// GetTokensPrices queries the remote APIs to get the token prices of all the tokens having the given ids
func (c *CryptoCompareClient) GetTokensPrices(currency string, ids []string) ([]types.TokenPrice, error) {
	var resStruct struct {
		Raw map[string]map[string]MarketTicker
	}
	query := fmt.Sprintf("/data/pricemultifull?fsyms=%s&tsyms=%s", strings.Join(ids, ","), currency)
	err := c.queryCoinGecko(query, &resStruct)
	if err != nil {
		return nil, err
	}

	// return nil, nil
	return c.ConvertCoingeckoPrices(resStruct.Raw), nil
}

func (c *CryptoCompareClient) ConvertCoingeckoPrices(tokens map[string]map[string]MarketTicker) []types.TokenPrice {
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
func (c *CryptoCompareClient) GetCUDOSPrice(currency string) (string, error) {
	ids := []string{"CUDOS"}
	prices, err := c.GetTokensPrices(currency, ids)
	if err != nil {
		return "", err
	}
	price := fmt.Sprintf("%g", prices[0].Price)
	return price, err
}

// queryCoinGecko queries the CoinGecko APIs for the given endpoint
func (c *CryptoCompareClient) queryCoinGecko(endpoint string, ptr interface{}) error {
	req, err := http.NewRequest("GET", "https://min-api.cryptocompare.com"+endpoint, nil)

	if err != nil {
		return err
	}

	var apiKey string
	if c.useProdApiKey {
		apiKey = c.config.Config.CryptoCompareProdApiKey
	} else {
		apiKey = c.config.Config.CryptoCompareFreeApiKey
	}

	if apiKey != "" {
		log.Debug().Str("module", "crypto-compare").Msg("using api key")
		req.Header.Set("authorization", fmt.Sprintf("Apikey %s", c.useProdApiKey))
	} else {
		log.Debug().Str("module", "crypto-compare").Msg("no api key provided")
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	rateLimitRemainderHeader := resp.Header.Get("x-ratelimit-remaining")
	if rateLimitRemainderHeader == "" {
		log.Error().Str("module", "crypto-compare").Msg("no rate limit header found")
	}

	rateLimitRemainder, err := strconv.Atoi(rateLimitRemainderHeader)
	if err != nil {
		log.Error().Str("module", "crypto-compare").Msg("error while parsing rate limit header")
	}

	if rateLimitRemainder < 600000 {
		c.useProdApiKey = false
		log.Warn().Str("module", "crypto-compare").Msg("Switching to crypto-compare free api key")
	}

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
