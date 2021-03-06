package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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

type BithumbBid struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

type BithumbResponse struct {
	Status string `json:"status"`
	Data   struct {
		Timestamp       string       `json:timestamp`
		PaymentCurrency string       `json:payment_currency`
		OrderCurrency   string       `json:order_currency`
		Bids            []BithumbBid `json:"bids"`
		Asks						[]BithumbBid `json:"asks`
	} `json:"data"`
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
	if len(decoded.Tick.Ask) == 0 ||
		len(decoded.Tick.Ask) == 0 ||
		len(decoded.Tick.Bid) == 0 ||
		len(decoded.Tick.Bid) == 0 {
		c <- requestResult{
			exchange:   "huobikr",
			tradePrice: -1,
			askPrice:   -1,
			askVolume:  -1,
			bidPrice:   -1,
			bidVolume:  -1,
		}
	} else {
		c <- requestResult{
			exchange:   "huobikr",
			tradePrice: decoded.Tick.Close,
			askPrice:   decoded.Tick.Ask[0],
			askVolume:  decoded.Tick.Ask[1],
			bidPrice:   decoded.Tick.Bid[0],
			bidVolume:  decoded.Tick.Bid[1],
		}
	}
}

func bithumbETHPrice(c chan requestResult) {
	response, err := GetData("https://api.bithumb.com/public/orderbook/ETH_KRW")
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error in bithumbethprice")
	}
	var decoded BithumbResponse
	json.Unmarshal(response, &decoded)
	if len(decoded.Data.Bids) == 0 ||
		 len(decoded.Data.Asks) == 0 {
			c <- requestResult{
				 exchange: "bithumb",
				 tradePrice: -1,
				 askPrice:   -1,
				 askVolume:  -1,
				 bidPrice:   -1,
				 bidVolume:  -1,
			 }
		 } else {
			 f1, _ := strconv.ParseFloat(decoded.Data.Asks[0].Price, 64)
			 f2, _ := strconv.ParseFloat(decoded.Data.Asks[0].Quantity, 64)
			 f3, _ := strconv.ParseFloat(decoded.Data.Bids[0].Price, 64)
			 f4, _ := strconv.ParseFloat(decoded.Data.Bids[0].Quantity, 64)
			 c <- requestResult{
				exchange:   "bithumb",
				tradePrice: f1,
				askPrice:   f1,
				askVolume:  f2,
				bidPrice:   f3,
				bidVolume:  f4,
			}
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
	bithumb := val["bithumb"]
	upbit := val["upbit"]
	// huobi := val["huobikr"]

	diff := getDiffPercent(upbit, bithumb)
	var result TickerMessage

	if diff.actualPercent < 0 {
		// huobi sell, upbit buy
		result = generateTickerMessage(0, upbit, bithumb, diff)
	} else {
		// upbit sell, huobi buy
		result = generateTickerMessage(1, upbit, bithumb, diff)
	}

	if result.shouldSend {
		go sendTextToTelegramChat(1967491369, result.messageString)
		// go sendTextToTelegramChat(303250131, result.messageString)
	} else {
		fmt.Println("it's okay to chill")
	}
}

func getDiffPercent(upbit requestResult, bithumb requestResult) Comparison {
	percentDiff := (upbit.tradePrice - bithumb.tradePrice) / upbit.tradePrice * 100
	absPercent := math.Abs(percentDiff)
	absAmount := math.Abs(upbit.tradePrice - bithumb.tradePrice)

	return Comparison{
		actualPercent:   percentDiff,
		absolutePercent: absPercent,
		absoluteAmount:  absAmount,
	}
}

func generateTickerMessage(
	direction int,
	upbit requestResult,
	bithumb requestResult,
	diff Comparison) TickerMessage {

	// if requestResult has negative value: huobi array is empty, therefore prone to null pointer exception
	if bithumb.tradePrice == -1 {
		return TickerMessage{
			messageString: "Error: bithumb API sucks dick",
			shouldSend:    false,
		}
	}

	var b bytes.Buffer
	var shouldSend bool
	ac := accounting.Accounting{Symbol: "???", Precision: 0}
	timeStamp := time.Now()
	upbitPriceString := "upbit eth price: " + ac.FormatMoney(upbit.tradePrice) + "\n"
	bithumbPriceString := "bithumb eth price: " + ac.FormatMoney(bithumb.tradePrice) + "\n"
	priceDiffString := "diff: " + ac.FormatMoney(diff.absoluteAmount) + "\n"
	percentDiffString := fmt.Sprintf("%f", math.Round(diff.absolutePercent*100)/100) + "%" + "\n"

	// start wrting in bytes
	b.WriteString("\n----------------------------------\n")
	b.WriteString(timeStamp.String() + "\n\n")
	b.WriteString(upbitPriceString)
	b.WriteString(bithumbPriceString)
	b.WriteString(priceDiffString)
	b.WriteString(percentDiffString)

	switch direction {
	case 0:
		b.WriteString("\n\nbithumb SELL, upbit BUY\n\n")
		upbitAsk := "upbit ASK price: " + ac.FormatMoney(upbit.askPrice) + "\n"
		huobiBid := "bithumb BID price: " + ac.FormatMoney(bithumb.bidPrice) + "\n"
		b.WriteString(upbitAsk)
		b.WriteString(huobiBid)
		diff := bithumb.bidPrice - upbit.askPrice
		diffString := "Practical Spread: " + ac.FormatMoney(diff) + "\n"
		b.WriteString(diffString)
		shouldSend = diff > 20000 && bithumb.bidVolume > 0.01
	case 1:
		b.WriteString("\n\nupbit SELL, bithumb BUY\n\n")
		huobiAsk := "bithumb ASK price: " + ac.FormatMoney(bithumb.askPrice) + "\n"
		upbitBid := "upbit BID price: " + ac.FormatMoney(upbit.bidPrice) + "\n"
		b.WriteString(huobiAsk)
		b.WriteString(upbitBid)
		diff := upbit.bidPrice - bithumb.askPrice
		diffString := "Practical Spread: " + ac.FormatMoney(diff) + "\n"
		b.WriteString(diffString)
		shouldSend = diff > 20000 && bithumb.askVolume > 0.01
	}
	return TickerMessage{
		messageString: b.String(),
		shouldSend:    shouldSend,
	}
}
