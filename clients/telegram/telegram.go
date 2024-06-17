package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func New(token string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Client{Bot: bot}, nil
}

func (c *Client) ListenForUpdates() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := c.Bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		handleUpdate(update)
	}
	return nil
}

func (c *Client) handleUpdate(update tgbotapi.Update) {
	var userID int64
	if update.Message != nil {
		userID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.Message.Chat.ID
	}

	userState, exists := userStates[userID]
	if !exists {
		userState = &UserState{ChatID: userID}
		userStates[userID] = userState
	}

	if update.CallbackQuery != nil {
		handleSubscription(c, update, userState)
		return
	}

	switch userState.Step {
	case 0:
		handleInitialStep(c, update, userState)
	case 1:
		handlePhoneNumberStep(c, update, userState)
	case 2:
		handleVerificationCodeStep(c, update, userState)
	case 3:
		handleFullNameStep(c, update, userState)
	case 4:
		handleBirthDateStep(c, update, userState)
	case 5:
		handleEmailStep(c, update, userState)
	case 6:
		handleAuthorizationPhoneStep(c, update, userState)
	case 7:
		handleAuthorizationPasswordStep(c, update, userState)
	case 8:
		handleMainMenu(c, update, userState)
	case 9:
		handleMatchSelection(c, update, userState)
	case 10:
		handleReferralProgram(c, update, userState)
	case 11:
		handleEnterReferralCode(c, update, userState)
	case 12:
		handleSubscription(c, update, userState)
	}
}

func handleMainMenu(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Предстоящие матчи" {
		userState.Step = 9
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер матча из списка или нажмите 'Назад' для возврата:")
		msg.ReplyMarkup = getMatchSelectionKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Реферальная программа" {
		userState.Step = 10
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = getReferralProgramKeyboard()
		c.Bot.Send(msg)
	}
}

func getInitialKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Регистрация"),
		tgbotapi.NewKeyboardButton("Авторизация"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func getBackKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Назад"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func getMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Предстоящие матчи"),
		tgbotapi.NewKeyboardButton("Реферальная программа"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func getMatchSelectionKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Назад"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func getReferralProgramKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Посмотреть реферальный код"),
		tgbotapi.NewKeyboardButton("Ввести реферальный код"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(
		[]tgbotapi.KeyboardButton{buttons[0], buttons[1]},
		[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("Назад")},
	)
	return keyboard
}
