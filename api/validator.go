package api

import (
	"github.com/techschool/simplebank/util"

	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// Check supported currency
		return util.IsSupportedCurrency(currency)
	}
	return false
}