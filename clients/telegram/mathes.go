package telegram

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/machinebox/graphql"
)

// getMatchesForTeam выполняет GraphQL запрос для получения списка матчей для указанной команды.
func getMatchesForTeam(teamID string) ([]Match, error) {
	client := graphql.NewClient("https://api.test.fanzilla.app/graphql")

	// Установка заголовка Authorization
	req := graphql.NewRequest(`
		query {
			match {
				getList{
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
		}
	`)
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJhdF9oYXNoIjoiaVZWSjRFMEhGWUl4dlRVL1VnaGxGTkN2dmdMd3NWcnRrOXI4WlJtakl2VT0iLCJzdWIiOiJDaVJtWVdFNE1XUmtNQzFoTjJRekxURXdNMlV0T0dWa1lTMDVOek5tTUdabE16a3daakFTQkd4a1lYQT0iLCJhdWQiOiJhcHBsaWNhdGlvbnMiLCJpc3MiOiJleGFtcGxlLmNvbS9hdXRoIiwiZXhwIjoxNzE5NDE2MTAyfQ.cVnXhdcLWyYq6DZY29eg0jOWz0bYTue0Vep3kE5lMT8YAqFjt1UIyIg0BQBDE1MbW3G14sZZBSOklOQVD6hNoQ")

	var respData struct {
		Match struct {
			GetList struct {
				List  []Match
				Total int
				Limit int
				Page  int
			}
		}
	}

	err := client.Run(context.Background(), req, &respData)
	if err != nil {
		fmt.Printf("GraphQL error: %v\n", err)
		return nil, fmt.Errorf("произошла ошибка при получении списка матчей: %w", err)
	}

	return respData.Match.GetList.List, nil
}

// Match представляет структуру данных для матча.
type Match struct {
	ID          string
	Title       string
	Description string
	StartTime   string
	EndTime     string
	Venue       struct {
		Title    string
		Location string
	}
	Status    string
	EventType string
	Cover     struct {
		PublicLink string
	}
	Season struct {
		ID string
	}
	Team1 struct {
		ID    string
		Title string
	}
	Team2 struct {
		ID    string
		Title string
	}
	Stage struct {
		Title string
	}
	CreatedTime string
	UpdatedTime string
}

func handleMatchSelection(c *Client, update tgbotapi.Update, userState *UserState) {
	if update.Message.Text == "Назад" {
		userState.Step = 8
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = getMainMenuKeyboard()
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
	msg.ReplyMarkup = getMainMenuKeyboard()
	c.Bot.Send(msg)
}
