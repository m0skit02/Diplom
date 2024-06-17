package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

var botAuthTokens BotAuthTokens

// AuthorizeBot выполняет авторизацию бота и сохраняет токены
func AuthorizeBot() error {
	requestBody := map[string]string{
		"query": `mutation login {
           auth {
               login(login: "79518330037", password: "94829") {
                   accessToken
                   idToken
                   refreshToken
               }
           }
       }`,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("https://api.test.fanzilla.app/graphql", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to authorize bot")
	}

	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return err
	}

	botAuthTokens = BotAuthTokens{
		AccessToken:  authResponse.Data.Auth.Login.AccessToken,
		IDToken:      authResponse.Data.Auth.Login.IDToken,
		RefreshToken: authResponse.Data.Auth.Login.RefreshToken,
	}
	log.Printf("Bot authorized using IDToken: %s", botAuthTokens.IDToken)
	return nil
}

func init() {
	if err := AuthorizeBot(); err != nil {
		log.Fatalf("Failed to authorize bot on startup: %v", err)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			if err := AuthorizeBot(); err != nil {
				log.Printf("Failed to refresh bot token: %v", err)
			}
		}
	}()
}
