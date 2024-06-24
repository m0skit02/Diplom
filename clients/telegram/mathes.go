package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// getMatchesForTeam выполняет GraphQL запрос для получения списка матчей для указанной команды.
func getMatchesForTeam(userToken string) ([]Match, error) {
	query := `
	query {
		match {
			getList(filter: {teamIds: ["cd3daaded70e4c18a6f79b9290fe917c"]}) {
				list {
					id
					title
					description
					startTime
					endTime
					venue {
						title
						location
					}
					status
					eventType
					cover {
						publicLink
					}
					season {
						id
					}
					team1 {
						id
						title
					}
					team2 {
						id
						title
					}
					stage {
						title
					}
					createdTime
					updatedTime
				}
				total
				limit
				page
			}
		}
	}`

	requestBody := map[string]interface{}{
		"query": query,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshalling request body: %v", err)
		return nil, err
	}

	url := "https://api.test.fanzilla.app/graphql"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)

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

	var result struct {
		Data struct {
			Match struct {
				GetList struct {
					List []Match
				}
			}
		}
	}

	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Printf("Error parsing matches response: %v", err)
		return nil, err
	}

	return result.Data.Match.GetList.List, nil
}

// createOrder создаёт новый заказ для указанного пользователя.
func createOrder(userID, token string) (*CreateOrderResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := `
	mutation CreateOrder(
  $userID: ID!
) {
  
    order {
        create(data: {
      userId: $userID
    }) {
            id
            user { id login }
            status
            createdTime
            items { id title price }
            price
            priceWithDiscount
            visibleId
            appliedPromocode
            promocodeValidUntil
            notificationProperties {
                enableEmail
                enableSms
                overrideEmail
                overridePhone
            }
        }
    }
}`

	variables := map[string]interface{}{
		"userID": "22b95822-c681-103e-9562-ebc0d060c136",
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoib200S0s4YXJleHhqaEtFRkdQM2xSbkNSNVBleEZXMkQ3M1o4ZHFRbFh5ND0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzIwMDIzODE2fQ.QXT5b4fikgjJQSrpjfAdS8qvYyPc4LKojfo2-jD0FhfRKhBDDeuFYjpIVlv43uM8f10TMyA0MpaYrSx337-gLg")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CreateOrderResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// handleViewMatches показывает доступные матчи пользователю.
func handleViewMatches(c *Client, update tgbotapi.Update, userState *UserState) {
	matches, err := getMatchesForTeam(userState.IDToken)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка матчей: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	if len(matches) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На данный момент нет доступных матчей.")
		c.Bot.Send(msg)
		return
	}

	msgText := "Выберите матч, введя номер:\n"
	for i, match := range matches {
		msgText += fmt.Sprintf("%d. %s - %s\n", i+1, match.Title, match.StartTime)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	c.Bot.Send(msg)
	userState.Step = 9 // Переходим в состояние выбора матча
}

// getAvailablePlaces выполняет GraphQL запрос для получения доступных мест для указанного матча.
func getAvailablePlaces(eventId, token string) ([]Place, error) {
	query := `
	query getPlaces($eventId: ID!) {
        price {
            getPlaces (filter: {placeStatuses: ACTIVE, eventId: $eventId, tag: ONLINE}) {
                list {
                    place {
                        id
                    }
                }
            }
        }
    }`

	variables := map[string]interface{}{
		"eventId": eventId,
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

	url := "https://api.test.fanzilla.app/graphql"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoib200S0s4YXJleHhqaEtFRkdQM2xSbkNSNVBleEZXMkQ3M1o4ZHFRbFh5ND0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzIwMDIzODE2fQ.QXT5b4fikgjJQSrpjfAdS8qvYyPc4LKojfo2-jD0FhfRKhBDDeuFYjpIVlv43uM8f10TMyA0MpaYrSx337-gLg")

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

	// Логируем только длину ответа и количество мест, чтобы убрать ненужные данные
	log.Printf("Response length: %d bytes", len(responseBody))

	var result struct {
		Data struct {
			Price struct {
				GetPlaces struct {
					List []struct {
						Place Place
					} `json:"list"`
				}
			}
		}
	}

	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Printf("Error parsing places response: %v", err)
		return nil, err
	}

	places := make([]Place, len(result.Data.Price.GetPlaces.List))
	for i, item := range result.Data.Price.GetPlaces.List {
		places[i] = item.Place
	}

	log.Printf("Number of available places: %d", len(places))
	return places, nil
}

// applyPromocodeToOrder применяет промокод к указанному заказу.
func applyPromocodeToOrder(orderId, token string) (*ApplyPromocodeResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := `
	mutation ApplyPromocodeToOrder($orderId: ID!, $promocode: String!, $validUntil: Datetime!) {
		order {
			applyPromocode(orderId: $orderId, promocode: $promocode, until: $validUntil) {
				id
				status
				price
				priceWithDiscount
				promocodeValidUntil
				appliedPromocode
				additionalData { loyaltyAmount }
				items {
					id title price priceWithDiscount loyaltyUsed reservedUntil
				}
			}
		}
	}`

	promocode := "BOT"
	validUntil := time.Now().Add(48 * time.Hour).UTC().Format("2006-01-02T15:04:05Z")
	// Setting the validUntil to 24 hours from now in ISO 8601 format
	variables := map[string]interface{}{
		"orderId":    orderId,
		"promocode":  promocode,
		"validUntil": validUntil,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoib200S0s4YXJleHhqaEtFRkdQM2xSbkNSNVBleEZXMkQ3M1o4ZHFRbFh5ND0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzIwMDIzODE2fQ.QXT5b4fikgjJQSrpjfAdS8qvYyPc4LKojfo2-jD0FhfRKhBDDeuFYjpIVlv43uM8f10TMyA0MpaYrSx337-gLg")

	// Logging request headers
	log.Printf("Request Headers: %v", req.Header)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Request body: %s", string(reqBody))
	log.Printf("Response status: %s", resp.Status)
	log.Printf("Response body: %s", string(responseBody))

	var result ApplyPromocodeResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP Error Status: %d", resp.StatusCode)
		return nil, fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	return &result, nil
}

// setOrderNotification устанавливает уведомления для указанного заказа.
func setOrderNotification(orderId, token string) (*SetOrderNotificationResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	query := `
	mutation SetOrderNotification($orderId: ID!) {
		order {
			setOrderNotification(orderId: $orderId, data: {
				enableEmail: false,
				enableSms: true,
				overrideEmail: null,
				overridePhone: "null"
			}) {
				id
				notificationProperties {
					enableEmail
					enableSms
					overridePhone
					overrideEmail
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"orderId": orderId,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoiOGU0cDMxdmV6amNzaUk1QmhDc2w3cE5rbVFqK0tVSlJIdjFMZCs1YkJRRT0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzE5MjY1MTEyfQ.YKkR5RU6tgBdFvuOHrafrHSbQjzT78lXzX4RX1Il4Oz4RC_dn9dr7bCFPG42OYExO7-cfgbOuGFFnWOOZFC1-g")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SetOrderNotificationResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// makePaymentLink создаёт ссылку на оплату для указанного заказа.
func makePaymentLink(orderId, adapterName, overrideEmail, overridePhone, token string) (*MakePaymentLinkResponse, error) {
	url := "https://api.test.fanzilla.app/graphql"

	log.Print(adapterName)
	log.Print(overrideEmail)
	log.Print(overridePhone)

	query := `
	mutation MakePaymentLink($orderId: ID!, $adapterName: String!, $overrideEmail: String!, $overridePhone: String!) {
		payments {
			makePaymentLink(orderId: $orderId, adapterName: $adapterName, additionalData: {
				overrideEmail: $overrideEmail,
				overridePhone: $overridePhone
			}) {
				link
				expiredIn
			}
		}
	}`

	variables := map[string]interface{}{
		"orderId":       orderId,
		"adapterName":   adapterName,
		"overrideEmail": overrideEmail,
		"overridePhone": overridePhone,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoib200S0s4YXJleHhqaEtFRkdQM2xSbkNSNVBleEZXMkQ3M1o4ZHFRbFh5ND0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzIwMDIzODE2fQ.QXT5b4fikgjJQSrpjfAdS8qvYyPc4LKojfo2-jD0FhfRKhBDDeuFYjpIVlv43uM8f10TMyA0MpaYrSx337-gLg")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result MakePaymentLinkResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func handleMatchSelection(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 8
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = getMainMenuKeyboard()
		c.Bot.Send(msg)
		return
	}

	matchIndex, err := strconv.Atoi(update.Message.Text) // Преобразуем введенный номер в индекс
	if err != nil || matchIndex < 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите корректный номер матча.")
		c.Bot.Send(msg)
		return
	}

	matches, err := getMatchesForTeam(userState.IDToken) // Запрос матчей заново для актуальности
	if err != nil {
		log.Printf("Error getting matches: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении матчей: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	if matchIndex > len(matches) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер матча вне диапазона.")
		c.Bot.Send(msg)
		return
	}

	selectedMatch := matches[matchIndex-1]
	places, err := getAvailablePlaces(selectedMatch.ID, userState.IDToken) // Запрос мест
	if err != nil {
		log.Printf("Error getting available places: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении мест: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	if len(places) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет доступных мест для этого матча.")
		c.Bot.Send(msg)
		return
	}

	// Выбор случайного места
	rand.Seed(time.Now().UnixNano())
	selectedPlace := places[rand.Intn(len(places))]

	// Преобразование userState.ChatID в строку
	userID := strconv.FormatInt(userState.ChatID, 10)
	log.Print(userID, "----------------")
	// Создание заказа для пользователя
	orderResponse, err := createOrder(userID, userState.IDToken)
	log.Println(orderResponse)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при создании заказа: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	// Резервирование билета для пользователя
	_, err = reserveTicketByPlace(selectedMatch.ID, selectedPlace.ID, orderResponse.Data.Order.Create.ID)
	if err != nil {
		log.Printf("Error reserving ticket: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при резервировании билета: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	// Применение промокода к заказу
	promocodeResponse, err := applyPromocodeToOrder(orderResponse.Data.Order.Create.ID, userState.IDToken)
	if err != nil {
		log.Printf("Error applying promocode: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при применении промокода: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	// Установка уведомлений для заказа
	notificationResponse, err := setOrderNotification(orderResponse.Data.Order.Create.ID, userState.IDToken)
	if err != nil {
		log.Printf("Error setting order notification: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при установке уведомлений для заказа: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	// Создание ссылки на оплату
	paymentLinkResponse, err := makePaymentLink(orderResponse.Data.Order.Create.ID, "Tinkoff", "oleg.gusev12@mail.ru", "+79518330037", userState.IDToken)
	if err != nil {
		log.Printf("Error creating payment link: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при создании ссылки на оплату: "+err.Error())
		c.Bot.Send(msg)
		return
	}

	// Использование ответа для отправки сообщения пользователю
	msgText := fmt.Sprintf(
		"Билет на матч %s успешно зарезервирован!\nНомер заказа: %s\nМесто: %s\nСтадион: %s\nЛокация: %s\nПримененный промокод: %s\nЦена со скидкой: %v\nУведомления: Email - %t, SMS - %t\nСсылка на оплату: %s (действительна до: %s)",
		selectedMatch.Title,
		orderResponse.Data.Order.Create.ID,
		selectedPlace.ID,
		selectedMatch.Venue.Title,
		selectedMatch.Venue.Location,
		promocodeResponse.Data.Order.ApplyPromocode.AppliedPromocode,
		promocodeResponse.Data.Order.ApplyPromocode.PriceWithDiscount,
		notificationResponse.Data.Order.SetOrderNotification.NotificationProperties.EnableEmail,
		notificationResponse.Data.Order.SetOrderNotification.NotificationProperties.EnableSms,
		paymentLinkResponse.Data.Payments.MakePaymentLink.Link,
		paymentLinkResponse.Data.Payments.MakePaymentLink.ExpiredIn,
	)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	c.Bot.Send(msg)
}

func reserveTicketByPlace(eventId, placeId, orderId string) (*ReserveTicketResponse, error) {

	log.Print(orderId)
	log.Print(placeId)
	log.Print(eventId)
	url := "https://api.test.fanzilla.app/graphql"

	query := `
	mutation ReserveTicketByPlace(
	  $eventId: ID!
	  $placeIds: [ID!]!
	  $orderId: ID
	  $userId: ID
	){
	  ticket {
	    reserveByPlace(data: {
	      eventId: $eventId,
	      placeIds: $placeIds,
	      tag: ONLINE,
	      orderId: $orderId,
	      userId: $userId
	    },
	    ignoreAdminBlocking: false
	    ) {
	      id
	      status
	      place {
	        id
	        number
	        coordinates {
	          x
	          y
	        }
	        row {
	          number
	          sector {
	            title
	          }
	        }
	      }
	      venue {
	        id
	        title
	        description
	        location
	      }
	      order {
	        id
	        status
	        createdTime
	        items {
	          id
	          title
	          type
	          status
	          item {
	            ... on Ticket {
	              id
	              price
	              status
	              place {
	                id
	                number
	              }
	              venue {
	                id
	                title
	              }
	            }
	          }
	          price
	          priceWithDiscount
	        }
	      }
	      user {
	        id
	        login
	        visibleId
	      }
	    }
	  }
	}
	`

	variables := map[string]interface{}{
		"eventId":  eventId,
		"placeIds": []string{placeId}, // исправлено на массив строк
		"orderId":  orderId,
		"userId":   "22b95822-c681-103e-9562-ebc0d060c136",
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoib200S0s4YXJleHhqaEtFRkdQM2xSbkNSNVBleEZXMkQ3M1o4ZHFRbFh5ND0iLCJzdWIiOiJDaVF4TnpnNVl6ZzNOaTA0TWpRNUxURXdNMkl0T0dNMFppMHhZalEwWW1RMU56VTJNV1VTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzIwMDIzODE2fQ.QXT5b4fikgjJQSrpjfAdS8qvYyPc4LKojfo2-jD0FhfRKhBDDeuFYjpIVlv43uM8f10TMyA0MpaYrSx337-gLg")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ReserveTicketResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Printf("Error parsing reservation response: %v, Response: %s", err, string(responseBody))
		return nil, err
	}

	log.Printf("Reservation successful: %+v", result)
	return &result, nil
}
