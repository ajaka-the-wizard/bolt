package utils

import (
	"log/slog"
	"reflect"

	"github.com/gofiber/fiber/v3"
)

// GetLogger gets the request scopped logger from locals using the generic function GetLocal
func GetLogger(c fiber.Ctx) *slog.Logger {
	if logger, ok := GetLocal[*slog.Logger](c, "logger"); ok {
		return logger
	}
	return slog.Default()
}

// GetLocal is a generic function that fetches anything from fiber request locals.
func GetLocal[T any](c fiber.Ctx, key string) (T, bool) {
	var zero T
	val := c.Locals(key)
	if val == nil {
		return zero, false
	}
	typed, ok := val.(T)
	if !ok || iszero(val) {
		return zero, false
	}
	return typed, true
}

// iszero checks if a field is set or still it's zero value
func iszero(d any) bool {
	if d == nil {
		return true
	}
	rv := reflect.ValueOf(d)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	}
	return reflect.DeepEqual(d, reflect.Zero(rv.Type()).Interface())
}
