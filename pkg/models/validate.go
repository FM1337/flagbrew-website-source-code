package models

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func validateStruct(v interface{}) error {
	err := validate.Struct(v)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			panic(err)
		}
		return err
	}
	return nil
}
