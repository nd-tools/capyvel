package uuid

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type UUID [16]byte

// New creates a new instance of UUID.
func New() *UUID {
	newUUID := uuid.New()
	var uuidStruct UUID
	copy(uuidStruct[:], newUUID[:])
	return &uuidStruct
}

// FromString converts a string to a UUID.
// If the UUID is empty or invalid, it returns nil.
func FromString(uuidStr string) (*UUID, error) {
	if uuidStr == "" || uuidStr == "00000000-0000-0000-0000-000000000000" {
		return nil, nil
	}
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID string: %v", err)
	}
	uid := New()
	copy((*uid)[:], parsedUUID[:])
	return uid, nil
}

// StringToUUID converts a string to a UUID.
// It returns nil for an empty or default UUID.
func StringToUUID(uuidStr string) *UUID {
	if uuidStr == "" || uuidStr == "00000000-0000-0000-0000-000000000000" {
		return nil
	}
	parsedUUID, _ := uuid.Parse(uuidStr) // Ignoramos el error
	uid := New()
	copy((*uid)[:], parsedUUID[:])
	return uid
}

// IsValid checks if the given string is a valid UUID.
func IsValid(uuidStr string) error {
	if uuidStr == "" {
		return errors.New("UUID string is empty")
	}
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		return fmt.Errorf("invalid UUID: %v", err)
	}

	return nil
}

// Equal checks if two UUID values are equal.
// Returns true if both UUIDs are equal or both are nil.
// Returns false if either is nil or if they are not equal.
func Equal(uuid1, uuid2 *UUID) bool {
	if uuid1 == nil && uuid2 == nil {
		return true
	}
	if uuid1 == nil || uuid2 == nil {
		return false
	}
	for i := 0; i < len(uuid1); i++ {
		if (*uuid1)[i] != (*uuid2)[i] {
			return false
		}
	}
	return true
}

// Scan implements the Scanner interface to convert database values to a UUID.
func (u *UUID) Scan(value interface{}) error {
	reverse := func(b []byte) {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	switch v := value.(type) {
	case []byte:
		if len(v) != 16 {
			return errors.New("mssql: invalid UniqueIdentifier length")
		}
		var raw UUID
		copy(raw[:], v)
		reverse(raw[0:4])
		reverse(raw[4:6])
		reverse(raw[6:8])
		*u = raw
		return nil
	case string:
		if len(v) != 36 {
			return errors.New("mssql: invalid UniqueIdentifier string length")
		}
		b := []byte(v)
		for i, c := range b {
			if c == '-' {
				b = append(b[:i], b[i+1:]...)
			}
		}
		bytes, err := hex.DecodeString(string(b))
		if err != nil {
			return err
		}
		copy(u[:], bytes)
		return nil
	default:
		return fmt.Errorf("mssql: cannot convert %T to UniqueIdentifier", v)
	}
}

// Value implements the driver.Valuer interface to convert a UUID to a value that can be stored in the database.
func (u UUID) Value() (driver.Value, error) {
	reverse := func(b []byte) {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}
	if u == (UUID{}) {
		return nil, nil
	}
	raw := make([]byte, len(u))
	copy(raw, u[:])

	reverse(raw[0:4])
	reverse(raw[4:6])
	reverse(raw[6:8])

	return raw, nil
}

// String converts the UUID to its string representation.
func (u UUID) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// MarshalText converts a UniqueIdentifier to bytes corresponding to its hexadecimal representation.
func (u UUID) MarshalText() (text []byte, err error) {
	text = []byte(u.String())
	return
}

// UnmarshalJSON converts a string representation of a UniqueIdentifier to bytes.
func (u *UUID) UnmarshalJSON(b []byte) error {
	input := strings.Trim(string(b), `"`)
	bytes, err := hex.DecodeString(strings.Replace(input, "-", "", -1))
	if err != nil {
		return err
	}
	copy(u[:], bytes)
	return nil
}
