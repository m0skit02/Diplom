package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleReferralProgram(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 8
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = getMainMenuKeyboard()
		c.Bot.Send(msg)
		return
	} else if update.Message.Text == "Посмотреть реферальный код" {
		referralCode := "12345" // Здесь должен быть вызов функции для получения реферального кода
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш реферальный код для друзей: "+referralCode)
		msg.ReplyMarkup = getReferralProgramKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Ввести реферальный код" {
		userState.Step = 11
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите реферальный код друга:")
		msg.ReplyMarkup = getBackKeyboard()
		c.Bot.Send(msg)
	}
}

func handleEnterReferralCode(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 10
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в реферальную программу.")
		msg.ReplyMarkup = getReferralProgramKeyboard()
		c.Bot.Send(msg)
		return
	}
	referralCode := update.Message.Text // Здесь должен быть вызов функции для обработки введенного реферального кода
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы ввели реферальный код: "+referralCode)
	msg.ReplyMarkup = getReferralProgramKeyboard()
	c.Bot.Send(msg)
	userState.Step = 10
}
