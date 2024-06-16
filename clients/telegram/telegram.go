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
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		c.handleUpdate(update)
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
		c.handleCallback(update)
		return
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
	case 8:
		c.handleMainMenu(update, userState)
	case 9:
		c.handleMatchSelection(update, userState)
	case 10:
		c.handleReferralProgram(update, userState)
	case 11:
		c.handleEnterReferralCode(update, userState)
	case 12:
		c.handleSubscription(update, userState) // Добавлен case для handleSubscription
	}
}

func (c *Client) handleCallback(update tgbotapi.Update) {
	if update.CallbackQuery != nil && update.CallbackQuery.Data == "subscribed" {
		userID := update.CallbackQuery.Message.Chat.ID
		userState := userStates[userID]
		userState.IsSubscribed = true // Устанавливаем флаг подписки
		msg := tgbotapi.NewMessage(userID, "Спасибо за подписку!")
		msg.ReplyMarkup = c.getMainMenuKeyboard()
		c.Bot.Send(msg)
		userState.Step = 8
		c.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Подписка подтверждена!"))
	}
}

func (c *Client) handleInitialStep(update tgbotapi.Update, userState *UserState) {
	if update.Message != nil && update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = c.getInitialKeyboard()
		c.Bot.Send(msg)
	} else if update.Message != nil && update.Message.Text == "Регистрация" {
		userState.Step = 1
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш номер телефона:")
		msg.ReplyMarkup = c.getBackKeyboard()
		c.Bot.Send(msg)
	} else if update.Message != nil && update.Message.Text == "Авторизация" {
		userState.Step = 6
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш номер телефона для авторизации:")
		msg.ReplyMarkup = c.getBackKeyboard()
		c.Bot.Send(msg)
	} else if update.Message != nil && update.Message.Text == "Назад" {
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

func (c *Client) getMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Предстоящие матчи"),
		tgbotapi.NewKeyboardButton("Реферальная программа"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func (c *Client) getMatchSelectionKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Назад"),
	}
	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	return keyboard
}

func (c *Client) getReferralProgramKeyboard() tgbotapi.ReplyKeyboardMarkup {
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
	userState.Step = 8 // Переход в главное меню после завершения регистрации
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Регистрация завершена! Добро пожаловать в главное меню.")
	msg.ReplyMarkup = c.getMainMenuKeyboard()
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
	userState.Step = 8 // Переход в главное меню после успешной авторизации
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Авторизация успешна! Добро пожаловать в главное меню.")
	msg.ReplyMarkup = c.getMainMenuKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleMainMenu(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Предстоящие матчи" {
		userState.Step = 9
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер матча из списка или нажмите 'Назад' для возврата:")
		msg.ReplyMarkup = c.getMatchSelectionKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Реферальная программа" {
		userState.Step = 10
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = c.getReferralProgramKeyboard()
		c.Bot.Send(msg)
	}
}

func (c *Client) handleMatchSelection(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 8
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = c.getMainMenuKeyboard()
		c.Bot.Send(msg)
		return
	}
	matchID := update.Message.Text // Здесь должен быть вызов функции для получения билета по matchID
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы получили билет на матч с ID: "+matchID)
	c.Bot.Send(msg)

	if !userState.IsSubscribed {
		adMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для получения следующего билета подпишитесь на наш канал: @Toporchan_Bot")
		adMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Подписался", "subscribed"),
			),
		)
		c.Bot.Send(adMsg)
	}

	userState.Step = 8
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в главное меню.")
	msg.ReplyMarkup = c.getMainMenuKeyboard()
	c.Bot.Send(msg)
}

func (c *Client) handleSubscription(update tgbotapi.Update, userState *UserState) {
	if update.CallbackQuery != nil && update.CallbackQuery.Data == "subscribed" {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Спасибо за подписку!")
		c.Bot.Send(msg)
		userState.IsSubscribed = true // Устанавливаем флаг подписки
		userState.Step = 8
		msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = c.getMainMenuKeyboard()
		c.Bot.Send(msg)
		c.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Подписка подтверждена!"))
	}
}

func (c *Client) handleReferralProgram(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 8
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = c.getMainMenuKeyboard()
		c.Bot.Send(msg)
		return
	} else if update.Message.Text == "Посмотреть реферальный код" {
		referralCode := "12345" // Здесь должен быть вызов функции для получения реферального кода
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш реферальный код для друзей: "+referralCode)
		msg.ReplyMarkup = c.getReferralProgramKeyboard()
		c.Bot.Send(msg)
	} else if update.Message.Text == "Ввести реферальный код" {
		userState.Step = 11
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите реферальный код друга:")
		msg.ReplyMarkup = c.getBackKeyboard()
		c.Bot.Send(msg)
	}
}

func (c *Client) handleEnterReferralCode(update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 10
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в реферальную программу.")
		msg.ReplyMarkup = c.getReferralProgramKeyboard()
		c.Bot.Send(msg)
		return
	}
	referralCode := update.Message.Text // Здесь должен быть вызов функции для обработки введенного реферального кода
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы ввели реферальный код: "+referralCode)
	msg.ReplyMarkup = c.getReferralProgramKeyboard()
	c.Bot.Send(msg)
	userState.Step = 10
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
	return inputPassword == "1234"
}
