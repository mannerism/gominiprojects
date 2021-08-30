package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mannerism/gominiprojects/util"
)

// https://medium.com/swlh/build-a-telegram-bot-in-go-in-9-minutes-e06ad38acef1
// https://www.educative.io/edpresso/how-to-make-a-bitcoin-telegram-bot-in-golang#hosting-it

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Text string `json:"test"`
	Chat Chat   `json:"chat"`
}

type Chat struct {
	Id int `json:"id"`
}

func parseTelegramRequest(r *http.Request) (*Update, error) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}
	return &update, nil
}

func processUserInput(w http.ResponseWriter, r *http.Request) {
	// Parse incoming request
	var update, err = parseTelegramRequest(r)
	if err != nil {
		log.Printf("error parsing update, %s", err.Error())
		return
	}
	var telegramResponseBody, errTelegram = sendTextToTelegramChat(update.Message.Chat.Id, "test")
	if errTelegram != nil {
		log.Printf("got error %s from telegram, response body is %s", errTelegram.Error(), telegramResponseBody)
	} else {
		log.Printf("message %s is successfully sent to chat id %d", "test", update.Message.Chat.Id)
	}

}

func sendTextToTelegramChat(chatId int, text string) (string, error) {
	log.Printf("Sending %s to chat_id: %d", text, chatId)

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	var telegramApi string = "https://api.telegram.org/bot" + config.TeleKey + "/sendMessage"
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})

	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}

	defer response.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error in parsing telegram answer %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram Response: %s", bodyString)
	return bodyString, nil

}
