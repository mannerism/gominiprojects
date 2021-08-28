package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/leekchan/accounting"
)

type UpbitResponse []struct {
	Market             string  `json:"market"`
	TradeDate          string  `json:"trade_date"`
	TradeTime          string  `json:"trade_time"`
	TradeDateKst       string  `json:"trade_date_kst"`
	TradeTimeKst       string  `json:"trade_time_kst"`
	TradeTimestamp     int64   `json:"trade_timestamp"`
	OpeningPrice       float64 `json:"opening_price"`
	HighPrice          float64 `json:"high_price"`
	LowPrice           float64 `json:"low_price"`
	TradePrice         float64 `json:"trade_price"`
	PrevClosingPrice   float64 `json:"prev_closing_price"`
	Change             string  `json:"change"`
	ChangePrice        float64 `json:"change_price"`
	ChangeRate         float64 `json:"change_rate"`
	SignedChangePrice  int     `json:"signed_change_price"`
	SignedChangeRate   float64 `json:"signed_change_rate"`
	TradeVolume        float64 `json:"trade_volume"`
	AccTradePrice      float64 `json:"acc_trade_price"`
	AccTradePrice24H   float64 `json:"acc_trade_price_24h"`
	AccTradeVolume     float64 `json:"acc_trade_volume"`
	AccTradeVolume24H  float64 `json:"acc_trade_volume_24h"`
	Highest52WeekPrice float64 `json:"highest_52_week_price"`
	Highest52WeekDate  string  `json:"highest_52_week_date"`
	Lowest52WeekPrice  float64 `json:"lowest_52_week_price"`
	Lowest52WeekDate   string  `json:"lowest_52_week_date"`
	Timestamp          int64   `json:"timestamp"`
}

type HuobiResponse struct {
	Status string `json:"status"`
	Ts     int64  `json:"ts"`
	Data   []struct {
		Symbol  string  `json:"symbol"`
		Open    float64 `json:"open"`
		High    float64 `json:"high"`
		Low     float64 `json:"low"`
		Close   float64 `json:"close"`
		Amount  float64 `json:"amount"`
		Vol     float64 `json:"vol"`
		Count   int     `json:"count"`
		Bid     float64 `json:"bid"`
		BidSize float64 `json:"bidSize"`
		Ask     float64 `json:"ask"`
		AskSize float64 `json:"askSize"`
	} `json:"data"`
}

func upbitETHPrice(c chan requestResult) {
	response, err := GetData("https://api.upbit.com/v1/ticker?markets=KRW-ETH")

	if err != nil {
		fmt.Println(err.Error())
	}

	var decoded UpbitResponse

	json.Unmarshal(response, &decoded)
	tradePrice := decoded[0].TradePrice

	c <- requestResult{exchange: "upbit", price: tradePrice}
}

func huobiETHPrice(c chan requestResult) {
	response, err := GetData("https://api-cloud.huobi.co.kr/market/tickers/")
	if err != nil {
		fmt.Println(err.Error())
	}

	var decoded HuobiResponse
	var tradePrice float64

	json.Unmarshal(response, &decoded)
	for _, single := range decoded.Data {
		if single.Symbol == "ethkrw" {
			tradePrice = single.Close
		}
	}
	c <- requestResult{exchange: "huobikr", price: tradePrice}
}

func priceChecker(val map[string]float64) {
	upbit := val["upbit"]
	huobi := val["huobikr"]
	percentDiff := (upbit - huobi) / upbit * 100
	abs := math.Abs(percentDiff)
	absdiff := math.Abs(upbit - huobi)
	ac := accounting.Accounting{Symbol: "₩", Precision: 0}
	fmt.Println("upbit eth price: ", ac.FormatMoney(upbit))
	fmt.Println("huobi eth price: ", ac.FormatMoney(huobi))
	fmt.Println("diff: ", ac.FormatMoney(absdiff))
	fmt.Println(math.Round(abs*100)/100, "%")

	if abs > 0.5 {
		fmt.Println("bro, you gotta check this shit")
	} else {
		fmt.Println("it's okay to chill")
	}
}
