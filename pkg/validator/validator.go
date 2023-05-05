package validator

import (
	"bank/pkg/apperror"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

var validate = validator.New()

func Validate(s any) error {
	err := validate.Struct(s)

	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			val := reflect.TypeOf(s)

			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}

			for _, validationError := range validationErrors {
				field, _ := val.FieldByName(validationError.Field())
				message := fmt.Sprintf("one of the specified parameters was missing or invalid: %s", prepareValidationError(field.Tag.Get("json"), validationError))
				return apperror.BadRequest.WithMessage(message)
			}
		}
	}

	return nil
}

func prepareValidationError(field string, err validator.FieldError) string {
	var addition string
	switch err.Kind() {
	case reflect.String:
		addition = " characters"
	}

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "gt":
		return fmt.Sprintf("%s must be longer than %s%s", field, err.Param(), addition)
	case "min", "gte":
		return fmt.Sprintf("%s must be longer or equal to %s%s", field, err.Param(), addition)
	case "lt":
		return fmt.Sprintf("%s cannot be longer than %s%s", field, err.Param(), addition)
	case "max", "lte":
		return fmt.Sprintf("%s cannot be longer or equal to %s%s", field, err.Param(), addition)
	case "oneof":
		return fmt.Sprintf("%s must be one of from: %s", field, strings.Replace(err.Param(), " ", ", ", -1))
	default:
		return fmt.Sprintf("%s is invalid, expected type %s", field, err.Tag())
	}
}
