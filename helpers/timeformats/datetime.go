package timeformats

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateTime para manejar fechas en formato ISO 8601
type DateTime struct {
	time.Time
}

const iso8601Layout = "2006-01-02T15:04:05Z07:00"

// UnmarshalJSON para DateTime
func (cdt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(iso8601Layout, s)
	if err != nil {
		return err
	}
	// Convertir de UTC a Local
	cdt.Time = t.UTC().Local()
	return nil
}

// MarshalJSON para DateTime
func (cdt DateTime) MarshalJSON() ([]byte, error) {
	// Convertir de Local a UTC antes de formatear
	utcTime := cdt.UTC()
	return []byte(fmt.Sprintf("\"%s\"", utcTime.Format(iso8601Layout))), nil
}

// Value para DateTime
func (cdt DateTime) Value() (driver.Value, error) {
	return cdt.UTC().Format(iso8601Layout), nil
}

// Scan para DateTime
func (cdt *DateTime) Scan(value interface{}) error {
	if value == nil {
		*cdt = DateTime{Time: time.Time{}}
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to scan DateTime: %v", value)
	}
	*cdt = DateTime{Time: t.UTC().Local()}
	return nil
}
