package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

func New(token string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Client{Bot: bot}, nil
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
		handleViewMatches(c, update, userState)
	case 10:
		handleMatchSelection(c, update, userState)
	default:
		handleUnknownCommand(c, update, userState)
	}
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
		c.handleUpdate(update)
	}
	return nil
}

func handleMainMenu(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Предстоящие матчи" {
		userState.Step = 9 // Переход к выбору матчей

		matches, err := getMatchesForTeam(userState.IDToken)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении списка матчей: "+err.Error())
			c.Bot.Send(msg)
			return
		}

		if len(matches) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Матчи не найдены.")
			c.Bot.Send(msg)
			return
		}

		responseText := "Список предстоящих матчей:\n"
		for i, match := range matches {
			startTime := formatDateTime(match.StartTime)
			matchInfo := fmt.Sprintf("%d. %s\n   %s - %s\n   %s (%s)\n\n",
				i+1,
				match.Title,
				startTime,
				match.Stage.Title,
				match.Venue.Title,
				match.Venue.Location)
			responseText += matchInfo
			if len(responseText) > 4000 { // Проверка на превышение максимального размера сообщения
				break
			}
		}
		fmt.Printf("Получено матчей: %d\n", len(matches))
		for i, match := range matches {
			fmt.Printf("Матч %d: %s\n", i+1, match.Title)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
		msg.ReplyMarkup = getBackKeyboard() // Предоставляем возможность вернуться назад
		c.Bot.Send(msg)

		userState.Step = 10 // Переход к выбору конкретного матча
	} else {
		handleUnknownCommand(c, update, userState)
	}
}

// Функция для форматирования даты и времени на русском языке
func formatDateTime(dateTimeStr string) string {
	t, err := time.Parse(time.RFC3339, dateTimeStr)
	if err != nil {
		return "неизвестно"
	}
	return t.Format("02.01.2006 в 15:04")
}

func getInitialKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Регистрация"),
			tgbotapi.NewKeyboardButton("Авторизация"),
		),
	)
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
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func (c *Client) handleSubscription(update tgbotapi.Update, userState *UserState) {
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

// Функция для обработки неверных команд
func handleUnknownCommand(c *Client, update tgbotapi.Update, userState *UserState) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда. Пожалуйста, выберите одну из доступных опций.")
	msg.ReplyMarkup = getMainMenuKeyboard() // Показываем главное меню
	c.Bot.Send(msg)
}
