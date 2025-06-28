package utils

import (
	"fmt"
	"regexp" // เพิ่ม import นี้สำหรับ regex
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func SetGlobalValidator(v *validator.Validate) {
	if v == nil {
		panic("Validator instance provided to SetGlobalValidator cannot be nil.")
	}
	validate = v
}

func GetGlobalValidator() *validator.Validate {
	if validate == nil {
		panic("Global validator has not been initialized. Call SetGlobalValidator() with a new validator.Validate() instance at application startup.")
	}
	return validate
}

func FormatValidationErrors(err error) map[string][]string {
	if err == nil {
		return nil
	}

	// Type assertion to get validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Not a validation error, return as a single generic error
		return map[string][]string{"_error_": {err.Error()}}
	}

	// Convert validator.ValidationErrors to our desired map[string][]string format
	formattedErrors := make(map[string][]string)
	for _, fieldError := range validationErrors {
		// Use Namespace() to get the path like "Categories.0.TrendsListNo"
		// And convert to snake_case for consistent JSON keys
		fieldName := toSnakeCase(fieldError.Namespace()) // Use Namespace() for full path
		fmt.Println(fieldName)

		// Generate a user-friendly error message based on the tag
		errorMessage := getErrorMessage(fieldError)

		formattedErrors[fieldName] = append(formattedErrors[fieldName], errorMessage)
	}

	return formattedErrors
}

// Regex to find uppercase letters that are not at the beginning of a word or after a period/bracket
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([A-Z])([A-Z][a-z]*)")
var matchNumAfterBracket = regexp.MustCompile(`\[(\d+)\]`) // Matches [number]

// toSnakeCase converts PascalCase/CamelCase field names to snake_case,
// handling array indices (e.g., "Categories[0].TrendsListNo" -> "categories.0.trends_list_no").
func toSnakeCase(s string) string {
	// Convert array access from "[0]" to ".0" first for easier processing
	// Example: "Categories[0].TrendsListNo" -> "Categories.0.TrendsListNo"
	s = matchNumAfterBracket.ReplaceAllString(s, ".$1")

	// Replace '.' with '_' temporarily to apply snake case globally, then revert
	tempS := strings.ReplaceAll(s, ".", "__DOT__")

	snake := matchFirstCap.ReplaceAllString(tempS, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	snake = strings.ToLower(snake)

	// Revert temporary '__DOT__' back to '.'
	snake = strings.ReplaceAll(snake, "__DOT__", ".")

	// Clean up any double underscores or underscores at the start/end
	snake = strings.ReplaceAll(snake, "__", "_")
	snake = strings.Trim(snake, "_")

	return snake
}

// getErrorMessage provides more readable error messages based on validation tag
func getErrorMessage(err validator.FieldError) string {
	// For error messages, using just Field() (e.g., "Name" instead of "user.name")
	// is often more user-friendly, but if you prefer the full path, use Namespace() here too.
	fieldName := toSnakeCase(err.Field()) // Use Field() for simple name, or Namespace() for full path
	param := err.Param()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldName)
	case "min":
		if err.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters long", fieldName, param)
		}
		return fmt.Sprintf("%s must be at least %s", fieldName, param)
	case "max":
		if err.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at most %s characters long", fieldName, param)
		}
		return fmt.Sprintf("%s must be at most %s", fieldName, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fieldName, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of %s", fieldName, strings.ReplaceAll(param, " ", ", "))
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fieldName)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldName)
	case "boolean":
		return fmt.Sprintf("%s must be a boolean", fieldName)
	case "numeric":
		return fmt.Sprintf("%s must be a number", fieldName)
	case "gte": // Greater than or equal
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldName, param)
	case "lte": // Less than or equal
		return fmt.Sprintf("%s must be less than or equal to %s", fieldName, param)
	case "gt": // Greater than
		return fmt.Sprintf("%s must be greater than %s", fieldName, param)
	case "lt": // Less than
		return fmt.Sprintf("%s must be less than %s", fieldName, param)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", fieldName)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", fieldName)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color code", fieldName)
	// Add more custom messages as needed for other tags
	default:
		// Fallback for tags not explicitly handled
		return fmt.Sprintf("Validation failed for %s on tag %s", fieldName, err.Tag())
	}
}
