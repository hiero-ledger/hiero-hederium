package rpc

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Register custom validators. Call this during server startup.

func RegisterCustomValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("eth_address", ethAddressValidator); err != nil {
			return err
		}

		if err := v.RegisterValidation("block_number_or_tag", blockNumberOrTagValidator); err != nil {
			return err
		}

		if err := v.RegisterValidation("hexadecimal", hexadecimalValidator); err != nil {
			return err
		}

		if err := v.RegisterValidation("eth_address_or_array", ethAddressOrArrayValidator); err != nil {
			return err
		}

		if err := v.RegisterValidation("startswith", startsWithValidator); err != nil {
			return err
		}

		return nil
	}
	return nil
}

// ethAddressValidator validates Ethereum addresses (0x followed by 40 hex chars)
func ethAddressValidator(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	return IsValidAddress(address)
}

// blockNumberOrTagValidator validates block numbers or special tags
func blockNumberOrTagValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return IsValidBlockNumberOrTag(value)
}

// hexadecimalValidator validates hexadecimal strings with 0x prefix
func hexadecimalValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return IsValidHexNumber(value)
}

// ethAddressOrArrayValidator validates either a single Ethereum address or an array of addresses
func ethAddressOrArrayValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return IsValidAddress(field.String())
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			if !IsValidAddress(field.Index(i).String()) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// startsWithValidator validates if a string starts with a specific prefix
func startsWithValidator(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	param := fl.Param()
	return strings.HasPrefix(field, param)
}
