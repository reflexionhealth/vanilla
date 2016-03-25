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

// Bool is a nullable boolean that doesn't require an extra allocation or dereference
// The builting sql package has a NullBool, but it doesn't implement json.Marshaler
type Bool sql.NullBool

func (nb *Bool) Set(value bool) {
	nb.Valid = true
	nb.Bool = value
}

// Implement sql.Scanner interface
func (nb *Bool) Scan(src interface{}) error {
	return (*sql.NullBool)(nb).Scan(src)
}

// Implement sql.driver.Valuer interface
func (nb Bool) Value() (driver.Value, error) {
	return (sql.NullBool)(nb).Value()
}

// Implement json.Marshaler interface
func (nb Bool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	} else {
		return []byte("null"), nil
	}
}

// Implement json.Unmarshaler interface
func (nb *Bool) UnmarshalJSON(bytes []byte) error {
	nb.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		nb.Bool = false
	} else {
		err := json.Unmarshal(bytes, &nb.Bool)
		if err != nil {
			return err
		} else {
			nb.Valid = true
		}
	}
	return nil
}

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
	ns.Valid = false
	if bytes == nil || string(bytes) == "null" {
		ns.String = ""
	} else {
		err := json.Unmarshal(bytes, &ns.String)
		if err != nil {
			return err
		} else {
			ns.Valid = true
		}
	}
	return nil
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
	nt.Valid = false
	if src == nil {
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
	nt.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		nt.Time = time.Time{}
	} else {
		err := nt.Time.UnmarshalJSON(bytes)
		if err != nil {
			return err
		} else {
			nt.Valid = true
		}
	}
	return nil
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
	nd.Valid = false
	if src == nil {
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

	nd.Valid = true
	nd.Date = date.From(nt.Time)
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
	nd.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		nd.Date = date.Date{}
	} else {
		err := nd.Date.UnmarshalJSON(bytes)
		if err != nil {
			return err
		} else {
			nd.Valid = true
		}
	}
	return nil
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
	id.Valid = false
	if src == nil {
		return nil
	}

	switch u := src.(type) {
	case string:
		var err error

		switch len(u) {
		case 32, 36:
			id.Uuid, err = uuid.FromString(u)
		case 16:
			id.Uuid, err = uuid.FromBytes([]byte(u))
		default:
			err = errors.New("sql/null: scan value for uuid was not 16, 32, or 36 bytes long")
		}

		if err != nil {
			return err
		}
	case []byte:
		var err error

		switch len(u) {
		case 32, 36:
			id.Uuid, err = uuid.FromString(string(u))
		case 16:
			id.Uuid, err = uuid.FromBytes(u)
		default:
			err = errors.New("sql/null: scan value for uuid was not 16, 32, or 36 bytes long")
		}

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

// Implement json.Unmarshaler interface
func (id *Uuid) UnmarshalJSON(bytes []byte) error {
	id.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		id.Uuid = uuid.UUID{} //date.Date{}
	} else {
		err := json.Unmarshal(bytes, &id.Uuid)
		if err != nil {
			return err
		} else {
			id.Valid = true
		}
	}
	return nil
}
