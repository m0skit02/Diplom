package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UserState struct {
	ChatID           int64
	Step             int
	PhoneNumber      string
	Password         string
	VerificationCode string
	Surname          string
	Name             string
	Patronymic       string
	BirthDate        string
	Email            string
	AccessToken      string
	IDToken          string
	RefreshToken     string
	IsSubscribed     bool
}

type Client struct {
	Bot *tgbotapi.BotAPI
}

type BotAuthTokens struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
}

var userStates = make(map[int64]*UserState)

type AuthResponse struct {
	Data struct {
		Auth struct {
			Login struct {
				AccessToken  string `json:"accessToken"`
				IDToken      string `json:"idToken"`
				RefreshToken string `json:"refreshToken"`
			} `json:"login"`
		} `json:"auth"`
	} `json:"data"`
}
