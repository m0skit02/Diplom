package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleSubscription(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.CallbackQuery != nil && update.CallbackQuery.Data == "subscribed" {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Спасибо за подписку!")
		c.Bot.Send(msg)
		userState.IsSubscribed = true // Устанавливаем флаг подписки
		userState.Step = 8
		msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = getMainMenuKeyboard()
		c.Bot.Send(msg)
		c.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Подписка подтверждена!"))
	}
}
