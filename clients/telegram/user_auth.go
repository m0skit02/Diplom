package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UserResponse struct {
	Data struct {
		Users struct {
			FindByLogin *struct {
				Login string `json:"login"`
			} `json:"findByLogin"`
		} `json:"users"`
	} `json:"data"`
}

// getUserPhoneNumbers отправляет запрос на сервер для получения информации о пользователе по номеру телефона
func getUserPhoneNumbers(phoneNumber string) (string, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := fmt.Sprintf(`{"query": "query{ users{ findByLogin(login: \"%s\"){ login } } }"}`, phoneNumber)

	reqBody, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botAuthTokens.IDToken) // Используем токен бота

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}

func checkPhoneNumber(phoneNumber string) (bool, error) {
	response, err := getUserPhoneNumbers(phoneNumber)
	if err != nil {
		log.Printf("Error retrieving user phone number: %v", err)
		return false, err
	}

	var userResponse UserResponse
	err = json.Unmarshal([]byte(response), &userResponse)
	if err != nil {
		log.Printf("Error parsing response: %v", err)
		return false, err
	}

	if userResponse.Data.Users.FindByLogin == nil {
		log.Printf("Phone number not found in DB: %s", phoneNumber)
		return false, nil
	}

	log.Printf("Phone number found in DB: %s", phoneNumber)
	return true, nil
}

func handleInitialStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message != nil && update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = getInitialKeyboard()
		c.Bot.Send(msg)
	} else if update.Message != nil && update.Message.Text == "Регистрация" {
		userState.Step = 1
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш номер телефона:")
		msg.ReplyMarkup = getBackKeyboard()
		c.Bot.Send(msg)
	} else if update.Message != nil && update.Message.Text == "Авторизация" {
		userState.Step = 6
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш номер телефона для авторизации:")
		msg.ReplyMarkup = getBackKeyboard()
		c.Bot.Send(msg)
	} else if update.Message != nil && update.Message.Text == "Назад" {
		userState.Step = 0
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись назад. Выберите действие:")
		msg.ReplyMarkup = getInitialKeyboard()
		c.Bot.Send(msg)
	}
}

func handlePhoneNumberStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
		return
	}

	if !validatePhoneNumber(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат номера. Пожалуйста, введите правильный номер:")
		c.Bot.Send(msg)
		return
	}
	userState.PhoneNumber = normalizePhoneNumber(update.Message.Text)
	userState.Step = 2

	valid, err := checkPhoneNumber(userState.PhoneNumber)
	if err != nil {
		log.Printf("Error checking phone number in DB: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при проверке номера. Попробуйте позже.")
		c.Bot.Send(msg)
		return
	}

	if !valid {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона не найден в базе данных. Пожалуйста, зарегистрируйтесь.")
		c.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите код, отправленный на ваш телефон:")
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)
}

func handleVerificationCodeStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
		return
	}

	if !validateVerificationCode(update.Message.Text, userState.VerificationCode) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный код подтверждения. Пожалуйста, попробуйте снова:")
		c.Bot.Send(msg)
		return
	}
	userState.Step = 3
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваше полное имя (Фамилия Имя Отчество):")
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)
}

func handleFullNameStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
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
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)
}

func handleBirthDateStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
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
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)
}

func handleEmailStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
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
	msg.ReplyMarkup = getMainMenuKeyboard()
	c.Bot.Send(msg)
}

func handleAuthorizationPhoneStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
		return
	}

	if !validatePhoneNumber(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат номера. Пожалуйста, введите правильный номер:")
		c.Bot.Send(msg)
		return
	}
	userState.PhoneNumber = normalizePhoneNumber(update.Message.Text)
	userState.Step = 7

	valid, err := checkPhoneNumber(userState.PhoneNumber)
	if err != nil {
		log.Printf("Error checking phone number in DB: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при проверке номера. Попробуйте позже.")
		c.Bot.Send(msg)
		return
	}

	if !valid {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона не найден в базе данных. Пожалуйста, зарегистрируйтесь.")
		c.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите пароль, отправленный на ваш телефон при регистрации:")
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)
}

func handleAuthorizationPasswordStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
		return
	}

	if !validatePassword(update.Message.Text, userState.PhoneNumber) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный пароль. Пожалуйста, попробуйте снова:")
		c.Bot.Send(msg)
		return
	}
	userState.Step = 8 // Переход в главное меню после успешной авторизации
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Авторизация успешна! Добро пожаловать в главное меню.")
	msg.ReplyMarkup = getMainMenuKeyboard()
	c.Bot.Send(msg)
}

func normalizePhoneNumber(phone string) string {
	if strings.HasPrefix(phone, "+7") {
		return phone[1:]
	} else if strings.HasPrefix(phone, "8") {
		return "7" + phone[1:]
	}
	return phone
}

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
