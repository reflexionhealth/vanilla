package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/reflexionhealth/vanilla/date"
	"github.com/satori/go.uuid"
)

var JsonNull = []byte("null")

// Bool is a nullable boolean that doesn't require an extra allocation or dereference.
// The builting sql package has a NullBool, but it doesn't implement json.Marshaler.
type Bool sql.NullBool

func (n *Bool) Set(value bool) {
	n.Valid = true
	n.Bool = value
}

// Implement sql.Scanner interface
func (n *Bool) Scan(src interface{}) error {
	return (*sql.NullBool)(n).Scan(src)
}

// Implement driver.Valuer interface
func (n Bool) Value() (driver.Value, error) {
	return (sql.NullBool)(n).Value()
}

// Implement json.Marshaler interface
func (n Bool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Bool)
	} else {
		return []byte("null"), nil
	}
}

// Implement json.Unmarshaler interface
func (n *Bool) UnmarshalJSON(bytes []byte) error {
	n.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		n.Bool = false
	} else {
		err := json.Unmarshal(bytes, &n.Bool)
		if err != nil {
			return err
		} else {
			n.Valid = true
		}
	}
	return nil
}

// String is a nullable string that doesn't require an extra allocation or dereference.
// The builting sql package has a NullString, but it doesn't implement json.Marshaler.
type String sql.NullString

func (n *String) Set(value string) {
	n.Valid = true
	n.String = value
}

// Implement sql.Scanner interface
func (n *String) Scan(src interface{}) error {
	return (*sql.NullString)(n).Scan(src)
}

// Implement driver.Valuer interface
func (n String) Value() (driver.Value, error) {
	return (sql.NullString)(n).Value()
}

// Implement json.Marshaler interface
func (n String) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	} else {
		return []byte("null"), nil
	}
}

// Implement json.Unmarshaler interface
func (n *String) UnmarshalJSON(bytes []byte) error {
	n.Valid = false
	if bytes == nil || string(bytes) == "null" {
		n.String = ""
	} else {
		err := json.Unmarshal(bytes, &n.String)
		if err != nil {
			return err
		} else {
			n.Valid = true
		}
	}
	return nil
}

type Float struct {
	Float float64
	Valid bool
}

func (n *Float) Set(value float64) {
	n.Valid = true
	n.Float = value
}

func (n *Float) Scan(src interface{}) error {
	n.Valid = false
	if src == nil {
		n.Float = 0.0
		return nil
	}

	switch t := src.(type) {
	case string:
		f64, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return fmt.Errorf("sql/null: converting driver.Value type %T (%q) to a null.Float: %v", src, t, strconvErr(err))
		}
		n.Set(f64)
	case float64:
		n.Set(t)
	case float32:
		n.Set(float64(t))
	case int64:
		n.Set(float64(t))
	case int32:
		n.Set(float64(t))
	}

	return nil
}

// Implement driver.Valuer interface
func (n Float) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	} else {
		return float64(n.Float), nil
	}
}

// Implement json.Marshaler interface
func (n Float) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Float)
	} else {
		return JsonNull, nil
	}
}

// Implement json.Unmarshaler interface
func (n *Float) UnmarshalJSON(bytes []byte) error {
	n.Valid = false
	if bytes == nil || string(bytes) == "null" {
		n.Float = 0.0
		return nil
	}

	err := json.Unmarshal(bytes, &n.Float)
	if err != nil {
		return err
	}

	n.Valid = true
	return nil
}

// Int is a nullable int that doesn't require an extra allocation or dereference.
// The builting sql package has a NullInt64, but it doesn't implement json.Marshaler
// and is an int64 instead of an int.
type Int struct {
	Int   int
	Valid bool
}

func (n *Int) Set(value int) {
	n.Valid = true
	n.Int = value
}

// Implement sql.Scanner interface
func (n *Int) Scan(src interface{}) error {
	n.Valid = false
	if src == nil {
		n.Int = 0
		return nil
	}
	switch t := src.(type) {
	case string:
		i64, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return fmt.Errorf("sql/null: converting driver.Value type %T (%q) to a null.Int: %v", src, t, strconvErr(err))
		}
		n.Set(int(i64))
	case int64:
		n.Set(int(t))
	}
	return nil
}

// Implement driver.Valuer interface
func (n Int) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	} else {
		return n.Int, nil
	}
}

// Implement json.Marshaler interface
func (n Int) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int)
	} else {
		return JsonNull, nil
	}
}

// Implement json.Unmarshaler interface
func (n *Int) UnmarshalJSON(bytes []byte) error {
	n.Valid = false
	if bytes == nil || string(bytes) == "null" {
		n.Int = 0
		return nil
	}

	err := json.Unmarshal(bytes, &n.Int)
	if err != nil {
		return err
	}

	n.Valid = true
	return nil
}

// Time is a nullable time.Time that doesn't require an extra allocation or dereference.
// It supports encoding/decoding with database/sql, encoding/gob, and encoding/json.
type Time struct {
	Time  time.Time
	Valid bool
}

func (n *Time) Set(value time.Time) {
	n.Valid = true
	n.Time = value
}

// Implement sql.Scanner interface
func (n *Time) Scan(src interface{}) error {
	n.Valid = false
	if src == nil {
		return nil
	}

	switch t := src.(type) {
	case string:
		var err error
		n.Time, err = time.Parse("2006-01-02 15:04:05", t)
		if err != nil {
			return err
		}
	case []byte:
		var err error
		n.Time, err = time.Parse("2006-01-02 15:04:05", string(t))
		if err != nil {
			return err
		}
	case time.Time:
		n.Time = t
	default:
		return errors.New("sql/null: scan value was not a Time, []byte, string, or nil")
	}

	n.Valid = true
	return nil
}

// Implement driver.Valuer interface
func (n Time) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	} else {
		return n.Time, nil
	}
}

// Implement json.Marshaler interface
func (n Time) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return n.Time.MarshalJSON()
	} else {
		return JsonNull, nil
	}
}

// Implement json.Unmarshaler interface
func (n *Time) UnmarshalJSON(bytes []byte) error {
	n.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		n.Time = time.Time{}
	} else {
		err := n.Time.UnmarshalJSON(bytes)
		if err != nil {
			return err
		} else {
			n.Valid = true
		}
	}
	return nil
}

// Date is a nullable date.Date that doesn't require an extra allocation or dereference.
// It supports encoding/decoding with database/sql, encoding/gob, and encoding/json.
type Date struct {
	Date  date.Date
	Valid bool
}

func (n *Date) Set(value date.Date) {
	n.Valid = true
	n.Date = value
}

// Implement sql.Scanner interface
func (n *Date) Scan(src interface{}) error {
	n.Valid = false
	if src == nil {
		return nil
	}

	var srcTime Time
	switch t := src.(type) {
	case string:
		var err error
		srcTime.Time, err = time.Parse("2006-01-02", t)
		if err != nil {
			return err
		}
	case []byte:
		var err error
		srcTime.Time, err = time.Parse("2006-01-02", string(t))
		if err != nil {
			return err
		}
	case time.Time:
		srcTime.Time = t
	default:
		return errors.New("sql/null: scan value was not a Time, []byte, string, or nil")
	}

	n.Valid = true
	n.Date = date.From(srcTime.Time)
	return nil
}

// Implement driver.Valuer interface
func (n Date) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	} else {
		return n.Date.Value()
	}
}

// Implement json.Marshaler interface
func (n Date) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return n.Date.MarshalJSON()
	} else {
		return JsonNull, nil
	}
}

// Implement json.Unmarshaler interface
func (n *Date) UnmarshalJSON(bytes []byte) error {
	n.Valid = false
	if bytes == nil || string(bytes) == `""` || string(bytes) == "null" {
		n.Date = date.Date{}
	} else {
		err := n.Date.UnmarshalJSON(bytes)
		if err != nil {
			return err
		} else {
			n.Valid = true
		}
	}
	return nil
}

// Uuid is a nullable date.Date that doesn't require an extra allocation or dereference.
// It supports encoding/decoding with database/sql, encoding/gob, and encoding/json.
type Uuid struct {
	Uuid  uuid.UUID
	Valid bool
}

func (id *Uuid) Set(value uuid.UUID) {
	id.Valid = true
	id.Uuid = value
}

// Implement sql.Scanner interface.
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

// Implement driver.Valuer interface
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

// copied from database/sql/convert.go
func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}
