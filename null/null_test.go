package null

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"testing"
	"time"

	"github.com/reflexionhealth/vanilla/date"
	"github.com/reflexionhealth/vanilla/uuid"

	"github.com/stretchr/testify/assert"
)

func TestImplementsJsonMarshaller(t *testing.T) {
	var marshaler json.Marshaler
	marshaler = Date{}
	assert.NotNil(t, marshaler)
	marshaler = Time{}
	assert.NotNil(t, marshaler)
	marshaler = String{}
	assert.NotNil(t, marshaler)
	marshaler = Int{}
	assert.NotNil(t, marshaler)
	marshaler = Bool{}
	assert.NotNil(t, marshaler)
	marshaler = UUID{}
	assert.NotNil(t, marshaler)
}

func TestImplementsJsonUnmarshaller(t *testing.T) {
	var unmarshaler json.Unmarshaler
	unmarshaler = &Date{}
	assert.NotNil(t, unmarshaler)
	unmarshaler = &Time{}
	assert.NotNil(t, unmarshaler)
	unmarshaler = &String{}
	assert.NotNil(t, unmarshaler)
	unmarshaler = &Int{}
	assert.NotNil(t, unmarshaler)
	unmarshaler = &Bool{}
	assert.NotNil(t, unmarshaler)
	unmarshaler = &UUID{}
	assert.NotNil(t, unmarshaler)
}

func TestImplementsSqlValuer(t *testing.T) {
	var valuer driver.Valuer
	valuer = Date{}
	assert.NotNil(t, valuer)
	valuer = Time{}
	assert.NotNil(t, valuer)
	valuer = String{}
	assert.NotNil(t, valuer)
	valuer = Int{}
	assert.NotNil(t, valuer)
	valuer = Bool{}
	assert.NotNil(t, valuer)
	valuer = UUID{}
	assert.NotNil(t, valuer)
}

func TestImplementSqlScanner(t *testing.T) {
	var scanner sql.Scanner
	scanner = &Date{}
	assert.NotNil(t, scanner)
	scanner = &Time{}
	assert.NotNil(t, scanner)
	scanner = &String{}
	assert.NotNil(t, scanner)
	scanner = &Int{}
	assert.NotNil(t, scanner)
	scanner = &Bool{}
	assert.NotNil(t, scanner)
	scanner = &UUID{}
	assert.NotNil(t, scanner)
}

func TestGobEncodeDecode(t *testing.T) {
	var buf bytes.Buffer
	// FIXME: preserve date timezone (or use UTC by default? see time.Time)
	var srcDate, destDate Date
	srcDate.Set(date.At(2033, 10, 24, nil))
	assert.Nil(t, gob.NewEncoder(&buf).Encode(srcDate))
	assert.Nil(t, gob.NewDecoder(&buf).Decode(&destDate))
	assert.Equal(t, srcDate, destDate)
	buf.Reset()

	var srcTime, destTime Time
	srcTime.Set(time.Now())
	assert.Nil(t, gob.NewEncoder(&buf).Encode(srcTime))
	assert.Nil(t, gob.NewDecoder(&buf).Decode(&destTime))
	assert.Equal(t, srcTime, destTime)
	buf.Reset()

	var srcString, destString String
	srcString.Set("gobify me")
	assert.Nil(t, gob.NewEncoder(&buf).Encode(srcString))
	assert.Nil(t, gob.NewDecoder(&buf).Decode(&destString))
	assert.Equal(t, srcString, destString)
	buf.Reset()

	var srcInt, destInt Int
	srcInt.Set(-154)
	assert.Nil(t, gob.NewEncoder(&buf).Encode(srcInt))
	assert.Nil(t, gob.NewDecoder(&buf).Decode(&destInt))
	assert.Equal(t, srcInt, destInt)
	buf.Reset()

	var srcBool, destBool Bool
	srcBool.Set(true)
	assert.Nil(t, gob.NewEncoder(&buf).Encode(srcBool))
	assert.Nil(t, gob.NewDecoder(&buf).Decode(&destBool))
	assert.Equal(t, srcBool, destBool)
	buf.Reset()

	var srcUUID, destUUID UUID
	srcUUID.Set(uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
	assert.Nil(t, gob.NewEncoder(&buf).Encode(srcUUID))
	assert.Nil(t, gob.NewDecoder(&buf).Decode(&destUUID))
	assert.Equal(t, srcUUID, destUUID)
	buf.Reset()
}

func TestSetNullable(t *testing.T) {
	var ns String
	assert.False(t, ns.Valid)
	ns.Set("hello world")
	assert.True(t, ns.Valid)

	var ni Int
	assert.False(t, ni.Valid)
	ni.Set(1)
	assert.True(t, ni.Valid)

	var nt Time
	assert.False(t, nt.Valid)
	nt.Set(time.Now())
	assert.True(t, nt.Valid)

	var nd Date
	assert.False(t, nd.Valid)
	nd.Set(date.From(time.Now()))
	assert.True(t, nd.Valid)

	var nb Bool
	assert.False(t, nb.Valid)
	nb.Set(true)
	assert.True(t, nb.Valid)

	var nu UUID
	assert.False(t, nu.Valid)
	nu.Set(uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
	assert.True(t, nu.Valid)
}

func TestUnmarshalNullBool(t *testing.T) {
	var jsonNull string = `null`
	var jsonEmpty string = `""`
	var bogusString string = `"bogus"`
	var validTrue string = `true`
	var validFalse string = `false`

	var n Bool
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(bogusString), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(validTrue), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)

	err = json.Unmarshal([]byte(validFalse), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
}

func TestUnmarshalNullInt(t *testing.T) {
	var jsonNull string = `null`
	var rationalFloat string = `12.22`
	var roundedFloat string = `16.0`
	var validZero string = `0`
	var validNegative string = `-300`
	var validPositive string = `1602525`

	var n Int
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(rationalFloat), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(roundedFloat), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(validZero), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, 0, n.Int)

	err = json.Unmarshal([]byte(validNegative), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, -300, n.Int)

	err = json.Unmarshal([]byte(validPositive), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, 1602525, n.Int)
}

func TestUnmarshalNullString(t *testing.T) {
	var jsonNull string = `null`
	var jsonNumber string = `3`
	var jsonEmpty string = `""`
	var validString string = `"foo"`

	var n String
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonNumber), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, "", n.String)

	err = json.Unmarshal([]byte(validString), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, "foo", n.String)

}

func TestUnmarshalNullTime(t *testing.T) {
	var jsonNull string = `null`
	var jsonEmpty string = `""`
	var stringTime string = `"2010-07-03T13:24:33Z"`
	var stringBogus string = `"bogus"`

	var n Time
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(stringTime), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, "2010-07-03 13:24:33", n.Time.Format("2006-01-02 15:04:05"))

	err = json.Unmarshal([]byte(stringBogus), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)
}

func TestUnmarshalNullDate(t *testing.T) {
	var jsonNull string = `null`
	var jsonEmpty string = `""`
	var stringDate string = `"2010-07-03"`
	var stringTime string = `"2010-07-03T13:24:33"`
	var stringBogus string = `"bogus"`

	var n Date
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	assert.Nil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(stringDate), &n)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, 2010, n.Date.Year)
	assert.Equal(t, time.July, n.Date.Month)
	assert.Equal(t, 3, n.Date.Day)

	err = json.Unmarshal([]byte(stringTime), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = json.Unmarshal([]byte(stringBogus), &n)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)
}

func TestScanNullTime(t *testing.T) {
	var rawTime = time.Now()
	var mysqlTime = "2010-07-03 13:24:33"
	var byteTime = []byte(mysqlTime)
	var notTime = 3

	var n Time
	var err error
	err = n.Scan(rawTime)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.NotEmpty(t, n.Time.Format("2006-01-02 15:04:05"))

	err = n.Scan(mysqlTime)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, mysqlTime, n.Time.Format("2006-01-02 15:04:05"))

	err = n.Scan(byteTime)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, mysqlTime, n.Time.Format("2006-01-02 15:04:05"))

	err = n.Scan(notTime)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)
}

func TestScanNullDate(t *testing.T) {
	var rawTime = time.Date(2010, time.July, 3, 13, 24, 33, 999, time.UTC)
	var mysqlTime = "2010-07-03 13:24:33"
	var mysqlDate = "2010-07-03"
	var byteTime = []byte(mysqlTime)
	var byteDate = []byte(mysqlDate)
	var notTime = 3

	var n Date
	var err error
	err = n.Scan(rawTime)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, 2010, n.Date.Year)
	assert.Equal(t, time.July, n.Date.Month)
	assert.Equal(t, 3, n.Date.Day)

	err = n.Scan(mysqlTime)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = n.Scan(mysqlDate)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, 2010, n.Date.Year)
	assert.Equal(t, time.July, n.Date.Month)
	assert.Equal(t, 3, n.Date.Day)

	err = n.Scan(byteTime)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)

	err = n.Scan(byteDate)
	assert.Nil(t, err)
	assert.True(t, n.Valid)
	assert.Equal(t, 2010, n.Date.Year)
	assert.Equal(t, time.July, n.Date.Month)
	assert.Equal(t, 3, n.Date.Day)

	err = n.Scan(notTime)
	assert.NotNil(t, err)
	assert.False(t, n.Valid)
}

func TestScanNullUUID(t *testing.T) {
	// start with a null UUID and scan a typical UUID
	{
		expectedUUID := uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
		stringUUID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

		n := UUID{}
		err := n.Scan(stringUUID)
		assert.Nil(t, err, "error unmarshaling null.UUID")
		assert.True(t, n.Valid, "null.UUID should be valid")
		assert.Equal(t, expectedUUID, n.UUID, "UUIDs should be equal")
	}

	// start with some UUID, and scan nil
	{
		n := SomeUUID(uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
		err := n.Scan(nil)
		assert.Nil(t, err, "error unmarshaling null.UUID")
		assert.False(t, n.Valid, "null.UUID should not be valid")
		assert.Equal(t, uuid.Nil, n.UUID, "null.UUID value should be equal to uuid.Nil")
	}
}

func TestValueNullUUID(t *testing.T) {
	u := UUID{}
	val, err := u.Value()
	assert.Nil(t, err, "error getting null.UUID value")
	assert.Nil(t, val, "wrong value returned, should be nil")
}
