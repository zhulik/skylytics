package clicfg

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/urfave/cli/v3"
)

var (
	ErrCannotParseFlags = errors.New("cannot parse flags")
)

func ParseFlags(c *cli.Command, s any) error {
	// Get the reflect.Value of the struct (must be a pointer)
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("%w: expected pointer to struct, got %T", ErrCannotParseFlags, s)
	}

	// Dereference the pointer to get the struct
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected pointer to struct, got pointer to %s", ErrCannotParseFlags, v.Kind())
	}

	// Get the struct type
	t := v.Type()

	// Iterate over all fields in the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the flag name from the tag
		flagName := field.Tag.Get("flag")
		if flagName == "" {
			continue
		}

		// Get the flag value from cli.Command based on field type
		switch field.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(c.String(flagName))
		case reflect.Bool:
			fieldValue.SetBool(c.Bool(flagName))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal := c.Int(flagName)
			fieldValue.SetInt(int64(intVal))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := c.Uint(flagName)
			fieldValue.SetUint(uint64(uintVal))
		case reflect.Float32, reflect.Float64:
			floatVal := c.Float64(flagName)
			fieldValue.SetFloat(floatVal)
		default:
			// For other types, try to get as string and convert
			strVal := c.String(flagName)
			if strVal != "" {
				if err := setValueFromString(fieldValue, strVal); err != nil {
					return fmt.Errorf("%w: failed to set field %s: %w", ErrCannotParseFlags, field.Name, err)
				}
			}
		}
	}

	return nil
}

// setValueFromString attempts to convert a string value to the target type
func setValueFromString(fieldValue reflect.Value, strVal string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(strVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(strVal)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)
	default:
		return fmt.Errorf("%w: unsupported type: %s", ErrCannotParseFlags, fieldValue.Kind())
	}
	return nil
}
