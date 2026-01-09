package validator

import (
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	currencies = []string{"USD", "EUR", "RUB"}
)

var (
	ValidCurrency validator.Func = func(fl validator.FieldLevel) bool {
		if currency, ok := fl.Field().Interface().(string); ok {
			return slices.Contains(currencies, currency)
		}
		return false
	}
	ValidPassword validator.Func = func(fl validator.FieldLevel) bool {
		if password, ok := fl.Field().Interface().(string); ok {
			startLen := len(password)
			if startLen >= 12 && startLen <= 38 {
				if len(strings.TrimSpace(password)) == startLen {
					return true
				}
				return false
			}
			return false
		}
		return false
	}
	ValidFullName validator.Func = func(fl validator.FieldLevel) bool {
		if fullName, ok := fl.Field().Interface().(string); ok {
			if len(strings.Split(fullName, " ")) >= 2 {
				return true
			}
			return false
		}
		return false
	}
)
