package steam

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

type MarketItemPriceOverview struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	MedianPrice string `json:"median_price"`
	Volume      uint32 `json:"volume,string"`
}

type MarketItemPrice struct {
	Date  string
	Price float64
	Count string
}

type MarketItemResponse struct {
	Success     bool        `json:"success"`
	PricePrefix string      `json:"price_prefix"`
	PriceSuffix string      `json:"price_suffix"`
	Prices      interface{} `json:"prices"`
}

var (
	ErrCannotLoadPrices     = errors.New("unable to load prices at this time")
	ErrInvalidPriceResponse = errors.New("invalid market pricehistory response")
)

func (community *Community) GetMarketItemPriceHistory(appID uint16, marketHashName string) ([]*MarketItemPrice, error) {
	resp, err := community.client.Get("https://steamcommunity.com/market/pricehistory/?" + url.Values{
		"appid":            {strconv.FormatUint(uint64(appID), 10)},
		"market_hash_name": {marketHashName},
	}.Encode())
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	response := MarketItemResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, ErrCannotLoadPrices
	}

	switch response.Prices.(type) {
	case []interface{}:
		items := []*MarketItemPrice{}
		for _, v := range response.Prices.([]interface{}) {
			switch v.(type) {
			case []interface{}:
				d := v.([]interface{})
				item := &MarketItemPrice{}
				for _, val := range d {
					switch val.(type) {
					case string:
						if len(item.Date) != 0 {
							item.Count = val.(string)
						} else {
							item.Date = val.(string)
						}
					case float64:
						item.Price = val.(float64)
					}
				}

				items = append(items, item)
			default:
				// ignore
			}
		}

		return items, nil
	case bool:
		return nil, ErrCannotLoadPrices
	}

	return nil, fmt.Errorf("GetMarketItemPriceHistory(): please implement type handler for %v", response.Prices)
}

func (community *Community) GetMarketItemPriceOverview(appID uint16, marketHashName string) (*MarketItemPriceOverview, error) {
	resp, err := community.client.Get("https://steamcommunity.com/market/priceoverview/?" + url.Values{
		"appid":            {strconv.FormatUint(uint64(appID), 10)},
		"market_hash_name": {marketHashName},
	}.Encode())
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	overview := &MarketItemPriceOverview{}
	if err = json.NewDecoder(resp.Body).Decode(overview); err != nil {
		return nil, err
	}

	return overview, nil
}
