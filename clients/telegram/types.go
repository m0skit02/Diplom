package telegram

type UserState struct {
	ChatID           int64
	PhoneNumber      string
	VerificationCode string
	Surname          string
	Name             string
	Patronymic       string
	BirthDate        string
	Email            string
	Step             int
	IsSubscribed     bool
}

var userStates = make(map[int64]*UserState)
