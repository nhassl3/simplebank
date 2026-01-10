package requests

type CreateAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

// CallAccountRequest uses in GetAccount, DeleteAccount
type CallAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type ListAccountsRequest struct {
	Page  int32 `form:"page" binding:"min=1"`
	Limit int32 `form:"limit" binding:"required,oneof=3 6 9 12"`
}

type UpdateAccountRequest struct {
	ID      int64 `json:"id" binding:"required,min=1"`
	Balance int64 `json:"balance" binding:"required,min=0"`
}

type AddAccountBalanceRequest struct {
	ID     int64 `json:"id" binding:"required,min=1"`
	Amount int64 `json:"amount" binding:"required"`
}

type TransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// User requests

// CallUserRequest uses in GetUser and DeleteUser methods
type CallUserRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,password"`
	FullName string `json:"full_name" binding:"required,fullname"`
	Email    string `json:"email" binding:"required,email"`
}

type UpdateUserPasswordRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,password"`
}

type UpdateUserFullNameRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	FullName string `json:"full_name" binding:"required,fullname"`
}
