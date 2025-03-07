package ptr

import "github.com/nd-tools/capyvel/helpers/uuid"

// Bool returns a pointer to the given bool value.
func Bool(b bool) *bool {
	return &b
}

// Int returns a pointer to the given int value.
func Int(i int) *int {
	return &i
}

// Int64 returns a pointer to the given int64 value.
func Int64(i int64) *int64 {
	return &i
}

// Float32 returns a pointer to the given float32 value.
func Float32(f float32) *float32 {
	return &f
}

// Float64 returns a pointer to the given float64 value.
func Float64(f float64) *float64 {
	return &f
}

// String returns a pointer to the given string value.
func String(s string) *string {
	return &s
}

// Uuid returns a pointer to the given uuid.UUID value.
func Uuid(u uuid.UUID) *uuid.UUID {
	return &u
}
