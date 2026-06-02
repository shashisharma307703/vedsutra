package types

import (
	"database/sql/driver"
	"strings"
	"time"
)

// NullDate handles flexible JSON parsing formats for postgres DATE columns
type NullDate struct {
	time.Time
	Valid bool
}

func (nd *NullDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		nd.Valid = false
		return nil
	}

	// 1. Try format YYYY-MM-DD
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		nd.Time = t
		nd.Valid = true
		return nil
	}

	// 2. Try full ISO timestamp format
	t, err = time.Parse("2006-01-02T15:04:05Z", s)
	if err == nil {
		nd.Time = t
		nd.Valid = true
		return nil
	}

	return err
}

func (nd NullDate) MarshalJSON() ([]byte, error) {
	if !nd.Valid {
		return []byte("null"), nil
	}
	return []byte(`"` + nd.Time.Format("2006-01-02") + `"`), nil
}

// Support pgx/v5 scanning and execution values driver assignments
func (nd NullDate) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	return nd.Time, nil
}

func (nd *NullDate) Scan(value interface{}) error {
	if value == nil {
		nd.Valid = false
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		nd.Valid = false
		return nil
	}
	nd.Time = t
	nd.Valid = true
	return nil
}