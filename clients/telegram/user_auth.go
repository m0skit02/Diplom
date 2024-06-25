package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getUserPhoneNumbers(phoneNumber string) (string, error) {
	url := "https://api.test.fanzilla.app/graphql"

	// Нормализация номера телефона для запроса
	normalizedPhone := normalizePhoneNumber(phoneNumber)

	query := `query findUserByLogin($login: String!) {
        users {
            findByLogin(login: $login) {
                login
            }
        }
    }`
	variables := map[string]interface{}{
		"login": normalizedPhone,
	}
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshalling request body: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoiRVZKNUVkaWd1OU41WnRGQmlkRTR2OHFJUk5LamdPMlN6US95d3ZoOWFWYz0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzE5NTI0MzEyfQ.YcKKN61dlmSMgKb9dlOUrGtmg2QUDW8auL5331XJ4Djyrg9ghkOST34TmLlbtFBDx0METlonBs4aItID7ZCQDA") // Replace with your actual token

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", err
	}

	log.Printf("GraphQL response: %s", responseBody)
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

	exists := userResponse.Data.Users.FindByLogin != nil
	log.Printf("Phone number %s found in DB: %v", phoneNumber, exists)
	return exists, nil
}

// Нормализует номер телефона, удаляя лишние символы и преобразуя в нужный формат.
func normalizePhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.NewReplacer("+", "", "-", "", " ", "").Replace(phone)
	if strings.HasPrefix(phone, "8") {
		phone = "7" + phone[1:]
	}
	return phone
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
		msg.ReplyMarkup = getInitialKeyboard() // Возвращаем пользователя к выбору регистрации или авторизации
		c.Bot.Send(msg)
		userState.Step = 0 // Возвращаем на начальный шаг
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

	if valid {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона уже зарегистрирован. Используйте другой номер или функцию восстановления пароля.")
		msg.ReplyMarkup = getInitialKeyboard() // Предоставляем выбор действий
		c.Bot.Send(msg)
		userState.Step = 0 // Возвращаем на начальный шаг
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите код, отправленный на ваш телефон:")
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)

	userState.CaptchaCode = "+7"

	log.Printf("Phone Number:%v, Captcha Code:%v", userState.PhoneNumber, userState.CaptchaCode)

	if _, err := registerUser(userState.PhoneNumber, userState.CaptchaCode); err != nil {
		log.Printf("Не удалось пройти регистрацию %v", err)
	}

	userState.Step = 2
}

func handleVerificationCodeStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
		return
	}

	if !validateVerificationCode(userState.PhoneNumber, update.Message.Text) {
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
	//fullName := userState.Surname + " " + userState.Name + " " + userState.Patronymic
	log.Println(Transliterate(userState.Surname))
	log.Println(Transliterate(userState.Name))
	log.Println(Transliterate(userState.Patronymic))
	resp, err := registerUser(userState.PhoneNumber, userState.CaptchaCode)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при регистрации. Попробуйте позже.")
		c.Bot.Send(msg)
		return
	}

	// Проверяем, что IDToken существует и не пуст
	if resp.Data.RegisterUserAndLogin.IDToken == "" {
		log.Printf("Received empty ID token during registration. IDToken: %v", resp.Data.RegisterUserAndLogin.IDToken)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка регистрации: не получен токен доступа.")
		c.Bot.Send(msg)
		return
	}

	// Сохранение IDToken для будущих запросов
	userState.IDToken = resp.Data.RegisterUserAndLogin.IDToken

	// Логирование для проверки присвоения токена
	log.Printf("User IDToken assigned after registration: %v", userState.IDToken)

	// Переход в главное меню после успешной регистрации
	userState.Step = 8
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
		msg.ReplyMarkup = getInitialKeyboard() // Возвращаем пользователя к выбору регистрации или авторизации
		c.Bot.Send(msg)
		userState.Step = 0 // Возвращаем на начальный шаг
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
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона не найден в базе данных. Пожалуйста, зарегистрируйтесь или попробуйте другой номер для авторизации.")
		msg.ReplyMarkup = getInitialKeyboard() // Предоставляем выбор действий
		c.Bot.Send(msg)
		userState.Step = 0 // Возвращаем на начальный шаг
		return
	}

	// Только если номер найден, разрешаем ввод пароля
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите пароль, отправленный на ваш телефон:")
	msg.ReplyMarkup = getBackKeyboard()
	c.Bot.Send(msg)
}

func handleAuthorizationPasswordStep(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		handleInitialStep(c, update, userState)
		return
	}

	authResponse, err := validatePassword(userState.PhoneNumber, update.Message.Text)
	if err != nil {
		if strings.Contains(err.Error(), "BAD_CREDENTIALS") {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введен неверный пароль. Пожалуйста, попробуйте еще раз или восстановите пароль, если забыли его.")
			msg.ReplyMarkup = getBackKeyboard() // Предоставляет пользователю возможность вернуться и попробовать снова
			c.Bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка авторизации: %v", err))
			msg.ReplyMarkup = getBackKeyboard()
			c.Bot.Send(msg)
		}
		return
	}

	// Добавляем логирование для отладки
	log.Printf("Auth response: %+v", authResponse)

	// Проверяем, что idToken существует и не пуст
	if authResponse.Data.Auth.Login.IDToken == "" {
		log.Printf("Received empty ID token. IDToken: %v", authResponse.Data.Auth.Login.IDToken)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка авторизации: не получен токен доступа.")
		c.Bot.Send(msg)
		return
	}

	// Сохранение idToken для будущих запросов
	userState.IDToken = authResponse.Data.Auth.Login.IDToken

	// Логирование для проверки присвоения токена
	log.Printf("User IDToken assigned: %v", userState.IDToken)

	// Переход в главное меню после успешной авторизации
	userState.Step = 8
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Авторизация успешна! Добро пожаловать в главное меню.")
	msg.ReplyMarkup = getMainMenuKeyboard()
	c.Bot.Send(msg)
}

func validatePhoneNumber(phone string) bool {
	re := regexp.MustCompile(`^\+7\d{10}$|^8\d{10}$`)
	return re.MatchString(phone)
}

func validateVerificationCode(phoneNumber string, inputCode string) bool {
	authResp, err := validatePassword(phoneNumber, inputCode)
	if err != nil {
		log.Printf("Не удалось пройти авторизаци: %v", err)
	}

	log.Printf("Ответ от авторизации %v", &authResp.Data.Auth.Login)

	if authResp.Data != nil {
		return true
	} else {
		return false
	}
	return false
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

func validatePassword(login, password string) (*AuthResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := `mutation Login($login: String!, $password: String!) {
        auth {
            login(login: $login, password: $password) {
                accessToken
                idToken
                refreshToken
            }
        }
    }`

	variables := map[string]interface{}{
		"login":    login,
		"password": password,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshalling request body: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var result AuthResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Printf("Error parsing authentication response: %v", err)
		return nil, err
	}

	log.Printf("Parsed response: %+v", result)
	if result.Data.Auth.Login.IDToken == "" {
		log.Printf("Received empty ID token. IDToken: %v", result.Data.Auth.Login.IDToken)
		return nil, fmt.Errorf("authentication error: ID token is empty")
	}

	return &result, nil
}

// registerUser функция для отправки запроса на регистрацию пользователя
func registerUser(phone, captchaCode string) (*RegistrationResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := `mutation RegisterUser($phone: String!, $captchaCode: String!) {
		registration {
			simpleUserRegistration(phone: $phone, captchaCode: $captchaCode) {
				login
			}
		}
	}`

	variables := map[string]interface{}{
		"phone":       phone,
		"captchaCode": captchaCode,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshalling request body: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var result RegistrationResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: Non-OK HTTP status: %d", resp.StatusCode)
		return nil, fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	return &result, nil
}
func Transliterate(input string) string {
	translitMap := map[rune]string{
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D",
		'Е': "E", 'Ё': "E", 'Ж': "Zh", 'З': "Z", 'И': "I",
		'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N",
		'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T",
		'У': "U", 'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch",
		'Ш': "Sh", 'Щ': "Shch", 'Ы': "Y", 'Э': "E", 'Ю': "Yu",
		'Я': "Ya", 'а': "a", 'б': "b", 'в': "v", 'г': "g",
		'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z",
		'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s",
		'т': "t", 'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts",
		'ч': "ch", 'ш': "sh", 'щ': "shch", 'ы': "y", 'э': "e",
		'ю': "yu", 'я': "ya", 'ь': "'",
	}

	var result strings.Builder
	for _, char := range input {
		if val, found := translitMap[char]; found {
			result.WriteString(val)
		} else {
			result.WriteRune(char)
		}
	}

	log.Println(result.String())

	return result.String()
}

func updateProfile(idToken, fullName, birthDate, phone, email string) (*UpdateProfileResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := `mutation UpdateUserProfile($user: ProfileUpdateInput!) {
		users {
			updateProfile(user: $user) {
				id
				person {
					name {
						ru
						en
					}
					surname {
						ru
						en
					}
					patronymic {
						ru
						en
					}
					birthday
					contacts {
						type
						value
					}
				}
			}
		}
	}`

	fullNameParts := strings.Split(fullName, " ")
	if len(fullNameParts) < 3 {
		return nil, fmt.Errorf("invalid fullName format, expected 'FirstName LastName Patronymic'")
	}

	name, surname, patronymic := fullNameParts[0], fullNameParts[1], fullNameParts[2]

	variables := map[string]interface{}{
		"user": map[string]interface{}{
			"password": "08704",
			"person": map[string]interface{}{
				"name": map[string]string{
					"ru": name,
					"en": Transliterate(name),
				},
				"surname": map[string]string{
					"ru": surname,
					"en": Transliterate(surname),
				},
				"patronymic": map[string]string{
					"ru": patronymic,
					"en": Transliterate(patronymic),
				},
				"birthday": birthDate,
				"contacts": []map[string]string{
					{"type": "PHONE", "value": phone},
					{"type": "EMAIL", "value": email},
				},
			},
		},
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshalling request body: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", idToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var result UpdateProfileResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: Non-OK HTTP status: %d", resp.StatusCode)
		return nil, fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	return &result, nil
}
