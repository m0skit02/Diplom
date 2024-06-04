package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"regexp"
	"strings"
)

type Client struct {
	Bot *tgbotapi.BotAPI
}

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
		if update.Message == nil {
			continue
		}
		c.handleUpdate(update)
	}
	return nil
}

func (c *Client) handleUpdate(update tgbotapi.Update) {
	userID := update.Message.Chat.ID
	userState, exists := userStates[userID]
	if !exists {
		userState = &UserState{ChatID: userID}
		userStates[userID] = userState
	}

	switch userState.Step {
	case 0:
		c.handleInitialStep(update, userState)
	case 1:
		c.handlePhoneNumberStep(update, userState)
	case 2:
		c.handleVerificationCodeStep(update, userState)
	case 3:
		c.handleFullNameStep(update, userState)
	case 4:
		c.handleBirthDateStep(update, userState)
	case 5:
		c.handleEmailStep(update, userState)
	case 6:
		c.handleAuthorizationPhoneStep(update, userState)
	case 7:
		c.handleAuthorizationPasswordStep(update, userState)
	}
}

func (c *Client) handleInitialStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = c.getInitialKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Регистрация" {
		userState.Step = 1
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш номер телефона:")
		msg.ReplyMarkup = c.getBackKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Авторизация" {
		userState.Step = 6
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш номер телефона для авторизации:")
		msg.ReplyMarkup = c.getBackKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Назад" {
		userState.Step = 0
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись назад. Выберите действие:")
		msg.ReplyMarkup = c.getInitialKeyboard()
		c.Bot.Send(msg)
	}
}

func (c *Client) getInitialKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Регистрация"),
		tgbotapi.NewKeyboardButton("Авторизация"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func (c *Client) getBackKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Назад"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func (c *Client) handlePhoneNumberStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	if !validatePhoneNumber(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат номера. Пожалуйста, введите правильный номер:")
		c.Bot.Send(msg)
		return
	}
	userState.PhoneNumber = update.Message.Text
	userState.Step = 2
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите код, отправленный на ваш телефон:")
	msg.ReplyMarkup = c.getBackKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleVerificationCodeStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	if !validateVerificationCode(update.Message.Text, userState.VerificationCode) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный код подтверждения. Пожалуйста, попробуйте снова:")
		c.Bot.Send(msg)
		return
	}
	userState.Step = 3
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваше полное имя (Фамилия Имя Отчество):")
	msg.ReplyMarkup = c.getBackKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleFullNameStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	surname, name, patronymic, ok := parseFullName(update.Message.Text)
	if !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите ваше полное имя правильно (Фамилия Имя Отчество):")
		c.Bot.Send(msg)
		return
	}
	userState.Surname = surname
	userState.Name = name
	userState.Patronymic = patronymic
	userState.Step = 4
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите вашу дату рождения (ДД.ММ.ГГГГ):")
	msg.ReplyMarkup = c.getBackKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleBirthDateStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	if !validateBirthDate(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат даты. Пожалуйста, введите дату в формате ДД.ММ.ГГГГ:")
		c.Bot.Send(msg)
		return
	}
	userState.BirthDate = update.Message.Text
	userState.Step = 5
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш email:")
	msg.ReplyMarkup = c.getBackKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleEmailStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	if !validateEmail(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат email. Пожалуйста, введите правильный email:")
		c.Bot.Send(msg)
		return
	}
	userState.Email = update.Message.Text
	userState.Step = 0                   // Сброс шага или установка шага завершения.
	delete(userStates, userState.ChatID) // Опционально удаляем состояние пользователя после завершения регистрации.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Регистрация завершена!")
	msg.ReplyMarkup = c.getInitialKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleAuthorizationPhoneStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	if !validatePhoneNumber(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат номера. Пожалуйста, введите правильный номер:")
		c.Bot.Send(msg)
		return
	}
	userState.PhoneNumber = update.Message.Text
	userState.Step = 7
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите пароль, отправленный на ваш телефон при регистрации:")
	msg.ReplyMarkup = c.getBackKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleAuthorizationPasswordStep(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		c.handleInitialStep(update, userState)
		return
	}

	if !validatePassword(update.Message.Text, userState.PhoneNumber) { // validatePassword - функция для проверки пароля
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный пароль. Пожалуйста, попробуйте снова:")
		c.Bot.Send(msg)
		return
	}
	userState.Step = 0                   // Сброс шага или установка шага завершения.
	delete(userStates, userState.ChatID) // Опционально удаляем состояние пользователя после завершения авторизации.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Авторизация успешна!")
	msg.ReplyMarkup = c.getInitialKeyboard()
	c.Bot.Send(msg)
}

// Функции валидации

func validatePhoneNumber(phone string) bool {
	re := regexp.MustCompile(`^\+7\d{10}$|^8\d{10}$`)
	return re.MatchString(phone)
}

func validateVerificationCode(inputCode, actualCode string) bool {
	actualCode = "1234" // Замените на реальный код в продакшене
	return inputCode == actualCode
}

func parseFullName(fullName string) (surname, name, patronymic string, ok bool) {
	parts := strings.Fields(fullName)
	switch len(parts) {
	case 2:
		return parts[0], parts[1], "", true
	case 3:
		return parts[0], parts[1], parts[2], true
	default:
		return "", "", "", false
	}
}

func validateBirthDate(birthDate string) bool {
	re := regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
	return re.MatchString(birthDate)
}

func validateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return re.MatchString(email)
}

func validatePassword(inputPassword, phoneNumber string) bool {
	// Замените на реальную проверку пароля
	return inputPassword == "password123"
}
