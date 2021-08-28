package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type requestResult struct {
	exchange string
	price    float64
}

func main() {
	doEvery(10*time.Second, startTicker)
}

func doEvery(d time.Duration, f func()) {
	for _ = range time.Tick(d) {
		f()
	}
}

func startTicker() {
	results := make(map[string]float64)

	exchanges := []string{
		"upbit",
		"huobikr",
	}

	c := make(chan requestResult)

	go upbitETHPrice(c)
	go huobiETHPrice(c)

	for i := 0; i < len(exchanges); i++ {
		result := <-c
		results[result.exchange] = result.price
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

	return responseData, err2
}
