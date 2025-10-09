package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidateStruct validates struct fields based on tags
func ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %T", s)
	}
	
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}
		
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			if err := validateField(field, rule, fieldType.Name); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func validateField(field reflect.Value, rule, fieldName string) error {
	switch rule {
	case "required":
		if field.Kind() == reflect.String && field.String() == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
	case "email":
		if field.Kind() == reflect.String {
			if !ValidateEmail(field.String()) {
				return fmt.Errorf("%s must be a valid email", fieldName)
			}
		}
	default:
		if strings.HasPrefix(rule, "min=") {
			minStr := strings.TrimPrefix(rule, "min=")
			if field.Kind() == reflect.String {
				if len(field.String()) < parseMinLength(minStr) {
					return fmt.Errorf("%s must be at least %s characters", fieldName, minStr)
				}
			}
		}
	}
	return nil
}

func parseMinLength(s string) int {
	if s == "8" {
		return 8
	}
	return 0
}