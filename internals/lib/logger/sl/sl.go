package sl

import (
	"log/slog"

	"github.com/pkg/errors"
)

// Custom errors
var (
	ErrorNoAccounts            = errors.New("failed to find accounts")
	ErrorAccountAlreadyExists  = errors.New("account already exists")
	ErrorFromAccountIDNotFound = errors.New("transfer sender not found")
	ErrorToAccountIDNotFound   = errors.New("transfer recipient not found")
	ErrorMismatchCurrencies    = errors.New("currencies do not match")
	ErrorBothAccountIDNotFound = errors.New("both accounts not found")
	ErrorNotEnoughMoney        = errors.New("sender does not have enough money")
	ErrorUserAlreadyExists     = errors.New("user already exists")
	ErrorNoUsers               = errors.New("no users found")
	ErrorPasswordsMatch        = errors.New("passwords match")
	ErrorUnauthorized          = errors.New("unauthorized")
	ErrorForbidden             = errors.New("operation forbidden for you")
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func ErrUpLevel(op string, err error) error {
	return errors.Wrap(err, op)
}
