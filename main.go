package TelegramBotFanzilla

import (
	"TelegramBotFanzilla/clients/telegram"
	"flag"
	"log"
)

func main() {

	tgClient := telegram.New(mustToken())

	// fetcher = fetcher.New(tgClient)

	// processor = processor.New(tgClient)

	// consumer.Start(fetcher, processor)
}

func mustToken() string {
	token := flag.String(
		"token-bot-token",
		"",
		"token for access to telegram bot",
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token
}
