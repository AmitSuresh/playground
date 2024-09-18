package validation

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidationError wraps the validator.FieldError so it is not exposed to our code
type ValidationError struct {
	FieldError validator.FieldError
	HasError   bool
}

func (v ValidationError) Error() string {
	return fmt.Sprintf(
		"Key: '%s' Error: Field validation for '%s' failed on the '%s' tag",
		v.FieldError.Namespace(),
		v.FieldError.Field(),
		v.FieldError.Tag(),
	)
}

// ValidationErrors is a collection of ValidationError
type ValidationErrors []ValidationError

// Errors converts the slice into a string slice
func (v ValidationErrors) Errors() []string {
	errs := []string{}
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

// CustomValidator contains a validate
type CustomValidator struct {
	validate *validator.Validate
}

// NewValidation creates a new CustomValidator type
func NewValidation() *CustomValidator {
	validate := validator.New()
	validate.RegisterValidation("sku", validateSKU)
	return &CustomValidator{validate}
}

// Validate the item
func (v *CustomValidator) Validate(i interface{}) ValidationErrors {
	errs := v.validate.Struct(i)
	if errs != nil {
		// Handle validation errors
		var returnErrs []ValidationError
		for _, err := range errs.(validator.ValidationErrors) {
			// cast the FieldError into our ValidationError and append to the slice
			ve := ValidationError{err, true}
			returnErrs = append(returnErrs, ve)
		}
		return returnErrs
	}
	return nil
}

// validateSKU
func validateSKU(fl validator.FieldLevel) bool {
	// SKU must be in the format abc-abc-abc
	re := regexp.MustCompile(`[a-z]+-[a-z]+-[a-z]+`)
	sku := re.FindAllString(fl.Field().String(), -1)

	return len(sku) == 1
}
