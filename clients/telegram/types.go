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
	FullName         string
	Surname          string
	Name             string
	Patronymic       string
	BirthDate        string
	Email            string
	AccessToken      string
	IDToken          string
	RefreshToken     string
	IsSubscribed     bool
	TeamID           string
	CaptchaCode      string
}

type Item struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Price             float64 `json:"price"`
	PriceWithDiscount float64 `json:"priceWithDiscount"`
	LoyaltyUsed       bool    `json:"loyaltyUsed"`
	ReservedUntil     string  `json:"reservedUntil"`
}

type CreateOrderResponse struct {
	Data struct {
		Order struct {
			Create struct {
				ID                     string  `json:"id"`
				User                   User    `json:"user"`
				Status                 string  `json:"status"`
				CreatedTime            string  `json:"createdTime"`
				Items                  []Item  `json:"items"`
				Price                  float64 `json:"price"`
				PriceWithDiscount      float64 `json:"priceWithDiscount"`
				VisibleID              string  `json:"visibleId"`
				AppliedPromocode       string  `json:"appliedPromocode"`
				PromocodeValidUntil    string  `json:"promocodeValidUntil"`
				NotificationProperties struct {
					EnableEmail   bool   `json:"enableEmail"`
					EnableSms     bool   `json:"enableSms"`
					OverrideEmail string `json:"overrideEmail"`
					OverridePhone string `json:"overridePhone"`
				} `json:"notificationProperties"`
			} `json:"create"`
		} `json:"order"`
	} `json:"data"`
}

// Структуры для резервирования билета
type ReserveTicketResponse struct {
	Data struct {
		Ticket struct {
			ReserveByPlace struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Place  struct {
					ID          string `json:"id"`
					Number      int    `json:"number"`
					Coordinates struct {
						X int `json:"x"`
						Y int `json:"y"`
					} `json:"coordinates"`
					Row struct {
						Number int `json:"number"`
						Sector struct {
							Title string `json:"title"`
						} `json:"sector"`
					} `json:"row"`
				} `json:"place"`
				Venue struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Location    string `json:"location"`
				} `json:"venue"`
				Order struct {
					ID          string `json:"id"`
					Status      string `json:"status"`
					CreatedTime string `json:"createdTime"`
					Items       []struct {
						ID     string `json:"id"`
						Title  string `json:"title"`
						Type   string `json:"type"`
						Status string `json:"status"`
						Item   struct {
							ID     string  `json:"id"`
							Price  float64 `json:"price"`
							Status string  `json:"status"`
							Place  struct {
								ID     string `json:"id"`
								Number int    `json:"number"`
							} `json:"place"`
							Venue struct {
								ID    string `json:"id"`
								Title string `json:"title"`
							} `json:"venue"`
						} `json:"item"`
						Price             float64 `json:"price"`
						PriceWithDiscount float64 `json:"priceWithDiscount"`
					} `json:"items"`
				} `json:"order"`
				User struct {
					ID        string `json:"id"`
					Login     string `json:"login"`
					VisibleID string `json:"visibleId"`
				} `json:"user"`
			} `json:"reserveByPlace"`
		} `json:"ticket"`
	} `json:"data"`
}

// Структуры для применения промокода
type ApplyPromocodeResponse struct {
	Data struct {
		Order struct {
			ApplyPromocode struct {
				ID                string  `json:"id"`
				Status            string  `json:"status"`
				Price             float64 `json:"price"`
				PriceWithDiscount float64 `json:"priceWithDiscount"`
				AppliedPromocode  string  `json:"appliedPromocode"`
				AdditionalData    struct {
					LoyaltyAmount float64 `json:"loyaltyAmount"`
				} `json:"additionalData"`
				Items []struct {
					ID                string  `json:"id"`
					Title             string  `json:"title"`
					Price             float64 `json:"price"`
					PriceWithDiscount float64 `json:"priceWithDiscount"`
					LoyaltyUsed       bool    `json:"loyaltyUsed"`
					ReservedUntil     string  `json:"reservedUntil"`
				} `json:"items"`
			} `json:"applyPromocode"`
		} `json:"order"`
	} `json:"data"`
}

// Структуры для установки уведомлений заказа
type SetOrderNotificationResponse struct {
	Data struct {
		Order struct {
			SetOrderNotification struct {
				ID                     string `json:"id"`
				NotificationProperties struct {
					EnableEmail   bool   `json:"enableEmail"`
					EnableSms     bool   `json:"enableSms"`
					OverrideEmail string `json:"overrideEmail"`
					OverridePhone string `json:"overridePhone"`
				} `json:"notificationProperties"`
			} `json:"setOrderNotification"`
		} `json:"order"`
	} `json:"data"`
}

// Структуры для создания ссылки на оплату
type MakePaymentLinkResponse struct {
	Data struct {
		Payments struct {
			MakePaymentLink struct {
				Link      string `json:"link"`
				ExpiredIn string `json:"expiredIn"`
			} `json:"makePaymentLink"`
		} `json:"payments"`
	} `json:"data"`
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

type UserResponse struct {
	Data struct {
		Users struct {
			FindByLogin *struct {
				Login string `json:"login"`
			} `json:"findByLogin"`
		} `json:"users"`
	} `json:"data"`
}

type RegisterUserResponse struct {
	Data struct {
		RegisterUser struct {
			ID        string `json:"id"`
			Login     string `json:"login"`
			Person    Person `json:"person"`
			Roles     []Role `json:"roles"`
			Deleted   string `json:"deletedDate"`
			Created   string `json:"createdDate"`
			Updated   string `json:"updatedDate"`
			VisibleID string `json:"visibleId"`
		} `json:"registerUser,omitempty"`
		RegisterUserAndLogin struct {
			AccessToken  string `json:"accessToken,omitempty"`
			IDToken      string `json:"idToken,omitempty"`
			RefreshToken string `json:"refreshToken,omitempty"`
		} `json:"registerUserAndLogin,omitempty"`
		SimpleUserRegistration struct {
			ID     string `json:"id"`
			Login  string `json:"login"`
			Person struct {
				Surname string `json:"surname"`
				Name    string `json:"name"`
			} `json:"person"`
		} `json:"simpleUserRegistration"`
	} `json:"data"`
}

type Person struct {
	Surname    string    `json:"surname"`
	Name       string    `json:"name"`
	Patronymic string    `json:"patronymic"`
	Birthday   string    `json:"birthday"`
	Contacts   []Contact `json:"contacts"`
}

type Contact struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Promocode содержит информацию о промокоде.
type Promocode struct {
	Value      string
	Descriptor struct {
		ID          string
		SectorBound struct {
			SectorIds []string
		}
	}
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

type Place struct {
	ID string `json:"id"`
}

type OrderResponse struct {
	Data struct {
		Order struct {
			Create struct {
				ID   string `json:"id"`
				User struct {
					ID    string `json:"id"`
					Login string `json:"login"`
				} `json:"user"`
				Status      string `json:"status"`
				CreatedTime string `json:"createdTime"`
				Items       []struct {
					ID    string  `json:"id"`
					Title string  `json:"title"`
					Price float64 `json:"price"`
				} `json:"items"`
				Price                  float64 `json:"price"`
				PriceWithDiscount      float64 `json:"priceWithDiscount"`
				VisibleID              string  `json:"visibleId"`
				AppliedPromocode       string  `json:"appliedPromocode"`
				PromocodeValidUntil    string  `json:"promocodeValidUntil"`
				NotificationProperties struct {
					EnableEmail   bool   `json:"enableEmail"`
					EnableSms     bool   `json:"enableSms"`
					OverrideEmail string `json:"overrideEmail"`
					OverridePhone string `json:"overridePhone"`
				} `json:"notificationProperties"`
			} `json:"create"`
		} `json:"order"`
	} `json:"data"`
}

// TicketReservationResponse структура для разбора ответа от сервера при резервировании билета
type TicketReservationResponse struct {
	Data struct {
		Ticket struct {
			ReserveByPlace struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Place  struct {
					ID          string `json:"id"`
					Number      string `json:"number"`
					Coordinates struct {
						X float64 `json:"x"`
						Y float64 `json:"y"`
					} `json:"coordinates"`
					Row struct {
						Number string `json:"number"`
						Sector struct {
							Title string `json:"title"`
						} `json:"sector"`
					} `json:"row"`
				} `json:"place"`
				Venue struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Location    string `json:"location"`
				} `json:"venue"`
				Order struct {
					ID          string `json:"id"`
					Status      string `json:"status"`
					CreatedTime string `json:"createdTime"`
					Items       []struct {
						ID     string `json:"id"`
						Title  string `json:"title"`
						Type   string `json:"type"`
						Status string `json:"status"`
						Item   struct {
							ID     string  `json:"id"`
							Price  float64 `json:"price"`
							Status string  `json:"status"`
							Place  struct {
								ID     string `json:"id"`
								Number string `json:"number"`
							} `json:"place"`
							Venue struct {
								ID    string `json:"id"`
								Title string `json:"title"`
							} `json:"venue"`
						} `json:"item"`
						Price             float64 `json:"price"`
						PriceWithDiscount float64 `json:"priceWithDiscount"`
					} `json:"items"`
				} `json:"order"`
				User struct {
					ID        string `json:"id"`
					Login     string `json:"login"`
					VisibleID string `json:"visibleId"`
				} `json:"user"`
			} `json:"reserveByPlace"`
		} `json:"ticket"`
	} `json:"data"`
}

type GetUserPhoneNumbersResponse struct {
	Data struct {
		User struct {
			PhoneNumbers []string `json:"phoneNumbers"`
		} `json:"user"`
	} `json:"data"`
}

type ValidatePasswordResponse struct {
	Data struct {
		ValidatePassword struct {
			IsValid bool `json:"isValid"`
		} `json:"validatePassword"`
	} `json:"data"`
}

type RegisterUserInput struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateProfileInput struct {
	UserID    string `json:"userID"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type UpdateProfileResponse struct {
	Data struct {
		UpdateProfile struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"updateProfile"`
	} `json:"data"`
}
