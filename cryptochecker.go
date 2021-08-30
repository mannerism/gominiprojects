package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"

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
	Ch     string `json:"ch"`
	Status string `json:"status"`
	Ts     int64  `json:"ts"`
	Tick   struct {
		ID      int64     `json:"id"`
		Version int64     `json:"version"`
		Open    float64   `json:"open"`
		Close   float64   `json:"close"`
		Low     float64   `json:"low"`
		High    float64   `json:"high"`
		Amount  float64   `json:"amount"`
		Vol     float64   `json:"vol"`
		Count   int       `json:"count"`
		Bid     []float64 `json:"bid"`
		Ask     []float64 `json:"ask"`
	} `json:"tick"`
}

func upbitETHPrice(c chan requestResult) {
	response, err := GetData("https://api.upbit.com/v1/ticker?markets=KRW-ETH")

	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error in upbitethprice")
	}

	var decoded UpbitResponse

	json.Unmarshal(response, &decoded)
	tradePrice := decoded[0].TradePrice

	c <- requestResult{
		exchange:   "upbit",
		tradePrice: tradePrice,
		askPrice:   tradePrice,
		askVolume:  0,
		bidPrice:   tradePrice,
		bidVolume:  0,
	}
}

func huobiETHPrice(c chan requestResult) {
	response, err := GetData("https://api-cloud.huobi.co.kr/market/detail/merged?symbol=ethkrw")
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error in huobiethprice")
	}

	var decoded HuobiResponse
	json.Unmarshal(response, &decoded)

	c <- requestResult{
		exchange:   "huobikr",
		tradePrice: decoded.Tick.Close,
		askPrice:   decoded.Tick.Ask[0],
		askVolume:  decoded.Tick.Ask[1],
		bidPrice:   decoded.Tick.Bid[0],
		bidVolume:  decoded.Tick.Bid[1],
	}
}

type Comparison struct {
	actualPercent   float64
	absolutePercent float64
	absoluteAmount  float64
}

type TickerMessage struct {
	messageString string
	shouldSend    bool
}

func priceChecker(val map[string]requestResult) {
	upbit := val["upbit"]
	huobi := val["huobikr"]
	diff := getDiffPercent(upbit, huobi)
	var result TickerMessage

	if diff.actualPercent < 0 {
		// huobi sell, upbit buy
		result = generateTickerMessage(0, upbit, huobi, diff)
	} else {
		// upbit sell, huobi buy
		result = generateTickerMessage(1, upbit, huobi, diff)
	}

	fmt.Println(result.messageString)

	if result.shouldSend {
		go sendTextToTelegramChat(1967491369, result.messageString)
		go sendTextToTelegramChat(303250131, result.messageString)
	} else {
		fmt.Println("it's okay to chill")
	}
}

func getDiffPercent(upbit requestResult, huobi requestResult) Comparison {
	percentDiff := (upbit.tradePrice - huobi.tradePrice) / upbit.tradePrice * 100
	absPercent := math.Abs(percentDiff)
	absAmount := math.Abs(upbit.tradePrice - huobi.tradePrice)

	return Comparison{
		actualPercent:   percentDiff,
		absolutePercent: absPercent,
		absoluteAmount:  absAmount,
	}
}

func generateTickerMessage(
	direction int,
	upbit requestResult,
	huobi requestResult,
	diff Comparison) TickerMessage {

	var b bytes.Buffer
	var shouldSend bool
	ac := accounting.Accounting{Symbol: "₩", Precision: 0}
	timeStamp := time.Now()
	upbitPriceString := "upbit eth price: " + ac.FormatMoney(upbit.tradePrice) + "\n"
	huobiPriceString := "huobi eth price: " + ac.FormatMoney(huobi.tradePrice) + "\n"
	priceDiffString := "diff: " + ac.FormatMoney(diff.absoluteAmount) + "\n"
	percentDiffString := fmt.Sprintf("%f", math.Round(diff.absolutePercent*100)/100) + "%" + "\n"

	// start wrting in bytes
	b.WriteString("\n----------------------------------\n")
	b.WriteString(timeStamp.String() + "\n\n")
	b.WriteString(upbitPriceString)
	b.WriteString(huobiPriceString)
	b.WriteString(priceDiffString)
	b.WriteString(percentDiffString)

	switch direction {
	case 0:
		b.WriteString("\n\nhuobi SELL, upbit BUY\n\n")
		upbitAsk := "upbit ASK price: " + ac.FormatMoney(upbit.askPrice) + "\n"
		huobiBid := "huobi BID price: " + ac.FormatMoney(huobi.bidPrice) + "\n"
		b.WriteString(upbitAsk)
		b.WriteString(huobiBid)
		diff := huobi.bidPrice - upbit.askPrice
		diffString := "Practical Spread: " + ac.FormatMoney(diff) + "\n"
		b.WriteString(diffString)
		shouldSend = diff > 8000 && huobi.bidVolume > 0.01
	case 1:
		b.WriteString("\n\nupbit SELL, huobi BUY\n\n")
		huobiAsk := "huobi ASK price: " + ac.FormatMoney(huobi.askPrice) + "\n"
		upbitBid := "upbit BID price: " + ac.FormatMoney(upbit.bidPrice) + "\n"
		b.WriteString(huobiAsk)
		b.WriteString(upbitBid)
		diff := upbit.bidPrice - huobi.askPrice
		diffString := "Practical Spread: " + ac.FormatMoney(diff) + "\n"
		b.WriteString(diffString)
		shouldSend = diff > 8000 && huobi.askVolume > 0.01
	}
	return TickerMessage{
		messageString: b.String(),
		shouldSend:    shouldSend,
	}
}
