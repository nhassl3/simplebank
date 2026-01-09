package simplebank

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/internals/http/simplebank/requests"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
	valid "github.com/nhassl3/simplebank/internals/lib/validator"
)

type Simplebank interface {
	CreateAccount(ctx context.Context, in requests.CreateAccountRequest) (*db.Account, error)
	GetAccount(ctx context.Context, id int64) (*db.Account, error)
	ListAccounts(ctx context.Context, in requests.ListAccountsRequest) (*[]db.Account, error)
	UpdateAccountBalance(ctx context.Context, in requests.UpdateAccountRequest) (*db.Account, error)
	AddAccountBalance(ctx context.Context, in requests.AddAccountBalanceRequest) (*db.Account, error)
	DeleteAccount(ctx context.Context, id int64) error
	CreateTransfer(ctx context.Context, in requests.TransferRequest) (*db.Transfer, error)
	CreateUser(ctx context.Context, in requests.CreateUserRequest) (*db.CreateUserRow, error)
	GetUser(ctx context.Context, username string) (*db.GetUserRow, error)
	UpdateUserPassword(ctx context.Context, in requests.UpdateUserPasswordRequest) (*db.UpdatePasswordRow, error)
	UpdateUserFullName(ctx context.Context, in requests.UpdateUserFullNameRequest) (*db.UpdateNameRow, error)
	DeleteUser(ctx context.Context, username string) error
}

type Server struct {
	simplebank Simplebank
	Router     *gin.Engine
}

func NewServer() *Server {
	var server Server

	// Initialize default router
	router := gin.Default()

	// Register new validation rule for struct tags in models or requests
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("currency", valid.ValidCurrency); err != nil {
			panic("failed to register validator for currency: " + err.Error())
		}
		if err := v.RegisterValidation("password", valid.ValidPassword); err != nil {
			panic("failed to register validator for password: " + err.Error())
		}
		if err := v.RegisterValidation("fullname", valid.ValidFullName); err != nil {
			panic("failed to register validator for fullname: " + err.Error())
		}
	}

	// REST API version 1:
	v1 := router.Group("/api/v1")
	{
		// Account endpoints set
		account := v1.Group("/accounts")
		{
			account.POST("/", server.CreateAccount)              // create account
			account.GET("/:id", server.GetAccount)               // get account by id in uri
			account.GET("/", server.ListAccounts)                // get accounts by query params
			account.PUT("/", server.UpdateAccountBalance)        // update balance
			account.PUT("/addBalance", server.AddAccountBalance) // add or remove balance
			account.DELETE("/:id", server.DeleteAccount)         // delete account
		}

		// Transfer endpoints set
		transfer := v1.Group("/transfers")
		{
			transfer.POST("/", server.CreateTransfer) // create transfer
		}

		// User endpoints set
		user := v1.Group("/users")
		{
			user.POST("/", server.CreateUser)
			user.GET("/:username", server.GetUser)
			user.PUT("/update/fullname", server.UpdateUserFullName)
			user.PUT("/update/password", server.UpdateUserPassword)
			user.DELETE("/:username", server.DeleteUser)
		}
	}

	server.Router = router

	return &server
}

func (s *Server) Register(simpleBankObj Simplebank) {
	s.simplebank = simpleBankObj
}

// CreateAccount creates new account with zero value balance and only for two currencies: USD and EUR
func (s *Server) CreateAccount(ctx *gin.Context) {
	var in requests.CreateAccountRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := s.simplebank.CreateAccount(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorAccountAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Account already exists"})
			return
		} else if errors.Is(err, sl.ErrorNoUsers) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "No users found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// GetAccount returns account with information by given ID
func (s *Server) GetAccount(ctx *gin.Context) {
	var in requests.CallAccountRequest
	if err := ctx.ShouldBindUri(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := s.simplebank.GetAccount(ctx, in.ID)
	if err != nil {
		if errors.Is(err, sl.ErrorNoAccounts) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// ListAccounts finds multiple accounts within the given limit and starts searching by the ID specified in the offset
func (s *Server) ListAccounts(ctx *gin.Context) {
	var in requests.ListAccountsRequest
	if err := ctx.ShouldBindQuery(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accounts, err := s.simplebank.ListAccounts(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorNoAccounts) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Accounts not found"})
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

// UpdateAccountBalance updates the account balance by replacing the number
func (s *Server) UpdateAccountBalance(ctx *gin.Context) {
	var in requests.UpdateAccountRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := s.simplebank.UpdateAccountBalance(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorNoAccounts) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// AddAccountBalance add the account balance by adding or subtracting the number
func (s *Server) AddAccountBalance(ctx *gin.Context) {
	var in requests.AddAccountBalanceRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := s.simplebank.AddAccountBalance(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorNoAccounts) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// DeleteAccount deletes account from the system but not delete user
func (s *Server) DeleteAccount(ctx *gin.Context) {
	var in requests.CallAccountRequest
	if err := ctx.ShouldBindUri(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.simplebank.DeleteAccount(ctx, in.ID); err != nil {
		if errors.Is(err, sl.ErrorNoAccounts) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": in.ID})
}

// CreateTransfer creates transaction with transfer doughs from one account to another
func (s *Server) CreateTransfer(ctx *gin.Context) {
	var in requests.TransferRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := s.simplebank.CreateTransfer(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorNotEnoughMoney) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Sender does not have enough money"})
			return
		} else if errors.Is(err, sl.ErrorBothAccountIDNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Accounts not found"})
			return
		} else if errors.Is(err, sl.ErrorFromAccountIDNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
			return
		} else if errors.Is(err, sl.ErrorToAccountIDNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
			return
		} else if errors.Is(err, sl.ErrorMismatchCurrencies) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Currencies do not match"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

// CreateUser creates new user in the system. User needed to create a few accounts
func (s *Server) CreateUser(ctx *gin.Context) {
	var in requests.CreateUserRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.simplebank.CreateUser(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorUserAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetUser returns user in the system by given username
func (s *Server) GetUser(ctx *gin.Context) {
	var in requests.CallUserRequest
	if err := ctx.ShouldBindUri(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.simplebank.GetUser(ctx, in.Username)
	if err != nil {
		if errors.Is(err, sl.ErrorNoUsers) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// UpdateUserPassword updates user password by given username. Password should be >= 12 with no spaces and <= 38
func (s *Server) UpdateUserPassword(ctx *gin.Context) {
	var in requests.UpdateUserPasswordRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.simplebank.UpdateUserPassword(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorNoUsers) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if errors.Is(err, sl.ErrorPasswordsMatch) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Passwords match"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// UpdateUserFullName updates the user's full name
func (s *Server) UpdateUserFullName(ctx *gin.Context) {
	var in requests.UpdateUserFullNameRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.simplebank.UpdateUserFullName(ctx, in)
	if err != nil {
		if errors.Is(err, sl.ErrorNoUsers) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// DeleteUser removes user from the system with him created accounts
func (s *Server) DeleteUser(ctx *gin.Context) {
	var in requests.CallUserRequest
	if err := ctx.ShouldBindUri(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.simplebank.DeleteUser(ctx, in.Username); err != nil {
		if errors.Is(err, sl.ErrorNoUsers) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
