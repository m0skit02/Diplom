package main

import (
	"TelegramBotFanzilla/clients/telegram"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	TelegramBotToken string `json:"telegram_bot_token"`
}

func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	bot, err := telegram.New(config.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	err = bot.ListenForUpdates()
	if err != nil {
		log.Panic(err)
	}
}
