package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type requestResult struct {
	exchange   string
	tradePrice float64
	askPrice   float64
	askVolume  float64
	bidPrice   float64
	bidVolume  float64
}

func main() {
	//	doEvery(5*time.Second, startTicker)
	bithumbETHPrice()
}

func doEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		f()
	}
}

func startTicker() {
	results := make(map[string]requestResult)

	exchanges := []string{
		"upbit",
		"huobikr",
	}

	c := make(chan requestResult)

	go upbitETHPrice(c)
	go huobiETHPrice(c)

	for i := 0; i < len(exchanges); i++ {
		result := <-c
		results[result.exchange] = result
	}
	priceChecker(results)

}

func GetData(url string) ([]byte, error) {
	response, err := http.Get(url)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	responseData, err2 := ioutil.ReadAll(response.Body)
	response.Body.Close()
	return responseData, err2
}
