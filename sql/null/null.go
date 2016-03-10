package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/reflexionhealth/vanilla/date"
	"github.com/satori/go.uuid"
)

var JsonNull = []byte("null")

// String is a nullable string that doesn't require an extra allocation or dereference
// The builting sql package has a NullString, but it doesn't implement json.Marshaler
type String sql.NullString

func (ns *String) Set(value string) {
	ns.Valid = true
	ns.String = value
}

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

// Implement json.Unmarshaler interface
func (ns *String) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		ns.Valid = false
		ns.String = ""
		return nil
	} else {
		ns.Valid = true
		err := json.Unmarshal(bytes, &ns.String)
		return err
	}
}

// Int64 is a nullable int64 that doesn't require an extra allocation or dereference
// The builting sql package has a NullInt64, but it doesn't implement json.Marshaler
type Int64 sql.NullInt64

func (ni *Int64) Set(value int64) {
	ni.Valid = true
	ni.Int64 = value
}

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

// Time is a nullable time.Time that doesn't require an extra allocation or dereference
type Time struct {
	Time  time.Time
	Valid bool
}

func (nt *Time) Set(value time.Time) {
	nt.Valid = true
	nt.Time = value
}

// Scan implements the sql.Scanner interface.
func (nt *Time) Scan(src interface{}) error {
	if src == nil {
		nt.Valid = false
		return nil
	}

	switch t := src.(type) {
	case string:
		var err error
		nt.Time, err = time.Parse("2006-01-02 15:04:05", t)
		if err != nil {
			return err
		}
	case []byte:
		var err error
		nt.Time, err = time.Parse("2006-01-02 15:04:05", string(t))
		if err != nil {
			return err
		}
	case time.Time:
		nt.Time = t
	default:
		return errors.New("sql/null: scan value was not a Time, []byte, string, or nil")
	}

	nt.Valid = true
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
		return nt.Time.MarshalJSON()
	} else {
		return JsonNull, nil
	}
}

// Implement json.Unmarshaler interface
func (nt *Time) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		nt.Valid = false
		nt.Time = time.Time{}
		return nil
	} else {
		nt.Valid = true
		err := nt.Time.UnmarshalJSON(bytes)
		return err
	}
}

// Date is a nullable date.Date that doesn't require an extra allocation or dereference
type Date struct {
	Date  date.Date
	Valid bool
}

func (nd *Date) Set(value date.Date) {
	nd.Valid = true
	nd.Date = value
}

// Implement sql.Scanner interface
func (nd *Date) Scan(src interface{}) error {
	if src == nil {
		nd.Valid = false
		return nil
	}

	var nt Time
	switch t := src.(type) {
	case string:
		var err error
		nt.Time, err = time.Parse("2006-01-02", t)
		if err != nil {
			return err
		}
	case []byte:
		var err error
		nt.Time, err = time.Parse("2006-01-02", string(t))
		if err != nil {
			return err
		}
	case time.Time:
		nt.Time = t
	default:
		return errors.New("sql/null: scan value was not a Time, []byte, string, or nil")
	}

	nd.Valid = nt.Valid
	if nt.Valid {
		nd.Date = date.From(nt.Time)
	}

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

// Implement json.Unmarshaler interface
func (nd *Date) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		nd.Valid = false
		nd.Date = date.Date{}
		return nil
	} else {
		nd.Valid = true
		err := nd.Date.UnmarshalJSON(bytes)
		return err
	}
}

// Uuid is a nullable date.Date that doesn't require an extra allocation or dereference
type Uuid struct {
	Uuid  uuid.UUID
	Valid bool
}

func (id *Uuid) Set(value uuid.UUID) {
	id.Valid = true
	id.Uuid = value
}

// Scan implements the sql.Scanner interface.
func (id *Uuid) Scan(src interface{}) error {
	if src == nil {
		id.Valid = false
		return nil
	}

	switch u := src.(type) {
	case string:
		var err error
		id.Uuid, err = uuid.FromString(u)
		if err != nil {
			return err
		}
	case []byte:
		var err error
		id.Uuid, err = uuid.FromString(string(u))
		if err != nil {
			return err
		}
	case uuid.UUID:
		id.Uuid = u
	default:
		return errors.New("sql/null: scan value was not a Time, []byte, string, or nil")
	}

	id.Valid = true
	return nil
}

// Implement sql.driver.Valuer interface
func (id Uuid) Value() (driver.Value, error) {
	if !id.Valid {
		return nil, nil
	} else {
		return id.Uuid.Value()
	}
}

// Implement json.Marshaler interface
func (id Uuid) MarshalJSON() ([]byte, error) {
	if id.Valid {
		return json.Marshal(id.Uuid)
	} else {
		return JsonNull, nil
	}
}
