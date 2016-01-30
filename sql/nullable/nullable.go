package nullable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/reflexionhealth/vanilla/date"
)

var JsonNull = []byte("null")

// String is a nullable string that doesn't require an extra allocation or dereference
// The builting sql package has a NullString, but it doesn't implement json.Marshaler
type String sql.NullString

// Implement sql.Scanner interface
func (ns *String) Scan(src interface{}) error {
	return (*sql.NullString)(ns).Scan(src)
}

// Implement sql.driver.Valuer interface
func (ns String) Value() (driver.Value, error) {
	return (sql.NullString)(ns).Value()
}

// Implement json.Marshaler interface
func (ns String) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	} else {
		return []byte("null"), nil
	}
}

// Int64 is a nullable int64 that doesn't require an extra allocation or dereference
// The builting sql package has a NullInt64, but it doesn't implement json.Marshaler
type Int64 sql.NullInt64

// Implement sql.Scanner interface
func (ni *Int64) Scan(src interface{}) error {
	return (*sql.NullInt64)(ni).Scan(src)
}

// Implement sql.driver.Valuer interface
func (ni Int64) Value() (driver.Value, error) {
	return (sql.NullInt64)(ni).Value()
}

// Implement json.Marshaler interface
func (ni Int64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	} else {
		return JsonNull, nil
	}
}

// Time represents a time.Time that doesn't require an extra allocation or dereference
type Time struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (nt *Time) Scan(src interface{}) error {
	if src == nil {
		nt.Valid = false
		return nil
	}

	t, ok := src.(time.Time)
	if !ok {
		return errors.New("sql/nullable: scan value for nullable.Time was not a Time or nil")
	}

	nt.Valid = true
	nt.Time = t
	return nil
}

// Value implements the sql.driver.Valuer interface
func (nt Time) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	} else {
		return nt.Time, nil
	}
}

// Implement json.Marshaler interface
func (nt Time) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time)
	} else {
		return JsonNull, nil
	}
}

// Date is a nullable Date that doesn't require an extra allocation or dereference
type Date struct {
	Date  date.Date
	Valid bool
}

// Implement sql.Scanner interface
func (nd *Date) Scan(src interface{}) error {
	if src == nil {
		nd.Valid = false
		return nil
	}

	t, ok := src.(time.Time)
	if !ok {
		return errors.New("sql/nullable: scan value for nullable.Date was not a time.Time or nil")
	}

	nd.Valid = true
	nd.Date = date.From(t)
	return nil
}

// Implement sql.driver.Valuer interface
func (nd Date) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	} else {
		return nd.Date.Value()
	}
}

// Implement json.Marshaler interface
func (nd Date) MarshalJSON() ([]byte, error) {
	if nd.Valid {
		return nd.Date.MarshalJSON()
	} else {
		return JsonNull, nil
	}
}