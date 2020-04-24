package shared

type RegisterUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RegisterUserResponse struct {
	UserID string `json:"userId"`
}

type CreateAccountRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccountType string `json:"type"`
}

type CreateAccountResponse struct {
	AccountID string `json:"accountId"`
}
