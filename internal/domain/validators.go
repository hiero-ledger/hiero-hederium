package domain

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	if err := RegisterCustomValidators(); err != nil {
		panic(fmt.Sprintf("Failed to register custom validators: %v", err))
	}
}

func RegisterCustomValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		customValidators := map[string]validator.Func{
			"eth_address":          ethAddressValidator,
			"block_number_or_tag":  blockNumberOrTagValidator,
			"hexadecimal":          hexadecimalValidator,
			"eth_address_or_array": ethAddressOrArrayValidator,
			"startswith":           startsWithValidator,
			"data":                 validateData,
		}

		for tag, validatorFn := range customValidators {
			if err := v.RegisterValidation(tag, validatorFn); err != nil {
				return err
			}
		}

		return nil
	}
	return nil
}

func translateValidationErrors(err error) (string, string) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {

		for _, e := range validationErrors {
			fieldName := e.Field()

			tag := e.Tag()
			value := e.Value()

			switch tag {
			case "eth_address":
				return fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", value), strings.ToLower(fieldName)
			case "block_number_or_tag":
				return fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", value), strings.ToLower(fieldName)
			case "hexadecimal":
				return fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", value), strings.ToLower(fieldName)
			case "eth_address_or_array":
				return fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", value), strings.ToLower(fieldName)
			case "startswith":
				return fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", value), strings.ToLower(fieldName)
			case "required":
				return fmt.Sprintf("Missing value for required parameter %s", fieldName), strings.ToLower(fieldName)
			case "data":
				return fmt.Sprintf("Expected 0x prefixed hexadecimal value with even length, value: %v", value), strings.ToLower(fieldName)
			default:
				return fmt.Sprintf("Field '%s' failed validation for '%s'", fieldName, tag), strings.ToLower(fieldName)
			}
		}

	}

	return err.Error(), ""
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

func validateData(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	return IsValidHexNumber(field) && len(field)%2 == 0
}
