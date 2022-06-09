package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func GetPriceFromCoinGecko(url string) (map[string]float64, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status err %d", rsp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) == 0 {
		return nil, fmt.Errorf("bodyBytes zero err")
	}
	coinGecko := RspCoinGecko{}
	err = json.Unmarshal(bodyBytes, &coinGecko)
	if err != nil {
		return nil, err
	}
	resPrice := make(map[string]float64)
	for key, value := range coinGecko {
		resPrice[key] = value.Usd
	}

	return resPrice, nil
}

func GetPriceFromCoinMarket(url string) (map[string]float64, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status err %d", rsp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) == 0 {
		return nil, fmt.Errorf("bodyBytes zero err")
	}
	coinMarket := RspCoinMarket{}
	err = json.Unmarshal(bodyBytes, &coinMarket)
	if err != nil {
		return nil, err
	}
	resPrice := make(map[string]float64)
	for key, v := range coinMarket.Data {
		resPrice[key] = v.Quote.USD.Price
	}
	return resPrice, nil
}

type RspCoinGecko map[string]CoinGeckeoPrice

type CoinGeckeoPrice struct {
	Usd float64 `json:"usd"`
}

type RspCoinMarket struct {
	Status struct {
		Timestamp    time.Time   `json:"timestamp"`
		ErrorCode    int         `json:"error_code"`
		ErrorMessage interface{} `json:"error_message"`
		Elapsed      int         `json:"elapsed"`
		CreditCount  int         `json:"credit_count"`
		Notice       interface{} `json:"notice"`
	} `json:"status"`

	Data map[string]TokenInfo `json:"data"`
}

type TokenInfo struct {
	ID                        int         `json:"id"`
	Name                      string      `json:"name"`
	Symbol                    string      `json:"symbol"`
	Slug                      string      `json:"slug"`
	NumMarketPairs            int         `json:"num_market_pairs"`
	DateAdded                 time.Time   `json:"date_added"`
	Tags                      []string    `json:"tags"`
	MaxSupply                 interface{} `json:"max_supply"`
	CirculatingSupply         float64     `json:"circulating_supply"`
	TotalSupply               float64     `json:"total_supply"`
	IsActive                  int         `json:"is_active"`
	IsMarketCapIncludedInCalc int         `json:"is_market_cap_included_in_calc"`
	Platform                  interface{} `json:"platform"`
	CmcRank                   int         `json:"cmc_rank"`
	IsFiat                    int         `json:"is_fiat"`
	LastUpdated               time.Time   `json:"last_updated"`
	Quote                     struct {
		USD struct {
			Price            float64   `json:"price"`
			Volume24H        float64   `json:"volume_24h"`
			PercentChange1H  float64   `json:"percent_change_1h"`
			PercentChange24H float64   `json:"percent_change_24h"`
			PercentChange7D  float64   `json:"percent_change_7d"`
			PercentChange30D float64   `json:"percent_change_30d"`
			PercentChange60D float64   `json:"percent_change_60d"`
			PercentChange90D float64   `json:"percent_change_90d"`
			MarketCap        float64   `json:"market_cap"`
			LastUpdated      time.Time `json:"last_updated"`
		} `json:"USD"`
	} `json:"quote"`
}
