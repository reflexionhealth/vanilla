package unstable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

var JsonNull = []byte("null")

// NullString is a nullable string that doesn't require an extra allocation or dereference
// The builting sql package has a NullString, but it doesn't implement json.Marshaler
type NullString sql.NullString

// Implement sql.Scanner interface
func (ns *NullString) Scan(src interface{}) error {
	return (*sql.NullString)(ns).Scan(src)
}

// Implement sql.driver.Valuer interface
func (ns NullString) Value() (driver.Value, error) {
	return (sql.NullString)(ns).Value()
}

// Implement json.Marshaler interface
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	} else {
		return []byte("null"), nil
	}
}

// NullInt64 is a nullable int64 that doesn't require an extra allocation or dereference
// The builting sql package has a NullInt64, but it doesn't implement json.Marshaler
type NullInt64 sql.NullInt64

// Implement sql.Scanner interface
func (ni *NullInt64) Scan(src interface{}) error {
	return (*sql.NullInt64)(ni).Scan(src)
}

// Implement sql.driver.Valuer interface
func (ni NullInt64) Value() (driver.Value, error) {
	return (sql.NullInt64)(ni).Value()
}

// Implement json.Marshaler interface
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	} else {
		return JsonNull, nil
	}
}

// NullTime represents a time.Time that doesn't require an extra allocation or dereference
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (nt *NullTime) Scan(src interface{}) error {
	if src == nil {
		nt.Valid = false
		return nil
	}

	t, ok := src.(time.Time)
	if !ok {
		return errors.New("unstable/nullable: scan value for NullTime was not a Time or nil")
	}

	nt.Valid = true
	nt.Time = t
	return nil
}

// Value implements the sql.driver.Valuer interface
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	} else {
		return nt.Time, nil
	}
}

// Implement json.Marshaler interface
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time)
	} else {
		return JsonNull, nil
	}
}

// NullDate is a nullable Date that doesn't require an extra allocation or dereference
type NullDate struct {
	Date  Date
	Valid bool
}

// Implement sql.Scanner interface
func (nd *NullDate) Scan(src interface{}) error {
	if src == nil {
		nd.Valid = false
		return nil
	}

	t, ok := src.(time.Time)
	if !ok {
		return errors.New("unstable/nullable: scan value for NullDate was not a Time or nil")
	}

	nd.Valid = true
	nd.Date = DateFrom(t)
	return nil
}

// Implement sql.driver.Valuer interface
func (nd NullDate) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	} else {
		return nd.Date.Value()
	}
}

// Implement json.Marshaler interface
func (nd NullDate) MarshalJSON() ([]byte, error) {
	if nd.Valid {
		return nd.Date.MarshalJSON()
	} else {
		return JsonNull, nil
	}
}
