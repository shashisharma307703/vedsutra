package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// CustomDate handles loose JSON timestamp parsing formats and maps natively to pgx/v5 Date fields
type CustomDate struct {
	pgtype.Date
}

// UnmarshalJSON intercepts JSON parsing for BOTH "2001-06-15" and "2001-06-15T00:00:00Z"
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		cd.Valid = false
		return nil
	}

	var parsedTime time.Time
	var err error

	// 1. Try format: Full ISO Timestamp
	if parsedTime, err = time.Parse("2006-01-02T15:04:05Z", s); err != nil {
		// 2. Try format: Simple ISO Date
		if parsedTime, err = time.Parse("2006-01-02", s); err != nil {
			return fmt.Errorf("failed to parse date input '%s': %w", s, err)
		}
	}

	cd.Time = parsedTime
	cd.Valid = true
	return nil
}

// MarshalJSON returns a clean date-only format string to frontends
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if !cd.Valid {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, cd.Time.Format("2006-01-02"))), nil
}

// Value tells database driver how to handle this type during writes
func (cd CustomDate) Value() (driver.Value, error) {
	if !cd.Valid {
		return nil, nil
	}
	return cd.Time, nil
}

// Scan tells database driver how to handle this type during reads
func (cd *CustomDate) Scan(src interface{}) error {
	if src == nil {
		cd.Valid = false
		return nil
	}
	
	switch t := src.(type) {
	case time.Time:
		cd.Time = t
		cd.Valid = true
		return nil
	default:
		// Fallback to native pgtype.Date scanner if format varies
		return cd.Date.Scan(src)
	}
}