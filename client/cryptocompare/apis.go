package cryptocompare

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

	"github.com/forbole/bdjuno/v4/types"
)

// NewModule returns a new Module instance
func NewClient(cfg *Config) *Client {
	return &Client{
		useProdAPIKey: true,
		config:        cfg,
	}
}

// GetTokensPrices queries the remote APIs to get the token prices of all the tokens having the given ids
func (c *Client) GetTokensPrices(currency string, ids []string) ([]types.TokenPrice, error) {
	var resStruct struct {
		Raw map[string]map[string]MarketTicker
	}
	query := fmt.Sprintf("/data/pricemultifull?fsyms=%s&tsyms=%s", strings.Join(ids, ","), currency)
	err := c.queryCryptoCompare(query, &resStruct)
	if err != nil {
		return nil, err
	}

	// return nil, nil
	return c.ConvertCoingeckoPrices(resStruct.Raw), nil
}

func (c *Client) ConvertCoingeckoPrices(tokens map[string]map[string]MarketTicker) []types.TokenPrice {
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
func (c *Client) GetCUDOSPrice(currency string) (string, error) {
	ids := []string{"CUDOS"}
	prices, err := c.GetTokensPrices(currency, ids)
	if err != nil {
		return "", err
	}
	price := fmt.Sprintf("%g", prices[0].Price)
	return price, err
}

// queryCryptoCompare queries the CoinGecko APIs for the given endpoint
func (c *Client) queryCryptoCompare(endpoint string, ptr interface{}) error {
	req, err := http.NewRequest("GET", "https://min-api.cryptocompare.com"+endpoint, nil)

	if err != nil {
		return err
	}

	var apiKey string
	var keyType = "empty"
	if c.useProdAPIKey {
		apiKey = c.config.Config.CryptoCompareProdAPIKey
	} else {
		apiKey = c.config.Config.CryptoCompareFreeAPIKey
	}

	if apiKey != "" {
		req.Header.Set("authorization", fmt.Sprintf("Apikey %s", apiKey))

		switch apiKey {
		case c.config.Config.CryptoCompareFreeAPIKey:
			keyType = "free"
		case c.config.Config.CryptoCompareProdAPIKey:
			keyType = "production"
		default:
			keyType = "no"
		}
	}

	log.Debug().Str("module", "crypto-compare").Msg(fmt.Sprintf("using %s api key", keyType))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	rateLimitRemainderHeader := resp.Header.Get("X-RateLimit-Remaining")
	if rateLimitRemainderHeader == "" {
		log.Warn().Str("module", "crypto-compare").Msg("no rate limit header found")
		rateLimitRemainderHeader = "0"
	}

	rateLimitRemainder, err := strconv.Atoi(rateLimitRemainderHeader)
	if err != nil {
		log.Warn().Str("module", "crypto-compare").Msg("error while parsing rate limit header")
		rateLimitRemainder = 0
	}

	if rateLimitRemainder < 600000 {
		c.useProdAPIKey = false
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
