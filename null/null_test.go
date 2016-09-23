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
	"github.com/reflexionhealth/vanilla/expect"
	"github.com/reflexionhealth/vanilla/uuid"
)

func TestImplementsJsonMarshaller(t *testing.T) {
	var marshaler json.Marshaler
	marshaler = Date{}
	expect.NotNil(t, marshaler)
	marshaler = Time{}
	expect.NotNil(t, marshaler)
	marshaler = String{}
	expect.NotNil(t, marshaler)
	marshaler = Int{}
	expect.NotNil(t, marshaler)
	marshaler = Bool{}
	expect.NotNil(t, marshaler)
	marshaler = UUID{}
	expect.NotNil(t, marshaler)
}

func TestImplementsJsonUnmarshaller(t *testing.T) {
	var unmarshaler json.Unmarshaler
	unmarshaler = &Date{}
	expect.NotNil(t, unmarshaler)
	unmarshaler = &Time{}
	expect.NotNil(t, unmarshaler)
	unmarshaler = &String{}
	expect.NotNil(t, unmarshaler)
	unmarshaler = &Int{}
	expect.NotNil(t, unmarshaler)
	unmarshaler = &Bool{}
	expect.NotNil(t, unmarshaler)
	unmarshaler = &UUID{}
	expect.NotNil(t, unmarshaler)
}

func TestImplementsSqlValuer(t *testing.T) {
	var valuer driver.Valuer
	valuer = Date{}
	expect.NotNil(t, valuer)
	valuer = Time{}
	expect.NotNil(t, valuer)
	valuer = String{}
	expect.NotNil(t, valuer)
	valuer = Int{}
	expect.NotNil(t, valuer)
	valuer = Bool{}
	expect.NotNil(t, valuer)
	valuer = UUID{}
	expect.NotNil(t, valuer)
}

func TestImplementSqlScanner(t *testing.T) {
	var scanner sql.Scanner
	scanner = &Date{}
	expect.NotNil(t, scanner)
	scanner = &Time{}
	expect.NotNil(t, scanner)
	scanner = &String{}
	expect.NotNil(t, scanner)
	scanner = &Int{}
	expect.NotNil(t, scanner)
	scanner = &Bool{}
	expect.NotNil(t, scanner)
	scanner = &UUID{}
	expect.NotNil(t, scanner)
}

func TestGobEncodeDecode(t *testing.T) {
	var buf bytes.Buffer
	// FIXME: preserve date timezone (or use UTC by default? see time.Time)
	var destDate, srcDate Date
	srcDate.Set(date.At(2033, 10, 24, nil))
	expect.Nil(t, gob.NewEncoder(&buf).Encode(srcDate))
	expect.Nil(t, gob.NewDecoder(&buf).Decode(&destDate))
	expect.Equal(t, destDate, srcDate)
	buf.Reset()

	var destTime, srcTime Time
	srcTime.Set(time.Now())
	expect.Nil(t, gob.NewEncoder(&buf).Encode(srcTime))
	expect.Nil(t, gob.NewDecoder(&buf).Decode(&destTime))
	expect.Equal(t, destTime, srcTime)
	buf.Reset()

	var destString, srcString String
	srcString.Set("gobify me")
	expect.Nil(t, gob.NewEncoder(&buf).Encode(srcString))
	expect.Nil(t, gob.NewDecoder(&buf).Decode(&destString))
	expect.Equal(t, destString, srcString)
	buf.Reset()

	var destInt, srcInt Int
	srcInt.Set(-154)
	expect.Nil(t, gob.NewEncoder(&buf).Encode(srcInt))
	expect.Nil(t, gob.NewDecoder(&buf).Decode(&destInt))
	expect.Equal(t, destInt, srcInt)
	buf.Reset()

	var destBool, srcBool Bool
	srcBool.Set(true)
	expect.Nil(t, gob.NewEncoder(&buf).Encode(srcBool))
	expect.Nil(t, gob.NewDecoder(&buf).Decode(&destBool))
	expect.Equal(t, destBool, srcBool)
	buf.Reset()

	var destUUID, srcUUID UUID
	srcUUID.Set(uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
	expect.Nil(t, gob.NewEncoder(&buf).Encode(srcUUID))
	expect.Nil(t, gob.NewDecoder(&buf).Decode(&destUUID))
	expect.Equal(t, destUUID, srcUUID)
	buf.Reset()
}

func TestSetNullable(t *testing.T) {
	var ns String
	expect.False(t, ns.Valid)
	ns.Set("hello world")
	expect.True(t, ns.Valid)

	var ni Int
	expect.False(t, ni.Valid)
	ni.Set(1)
	expect.True(t, ni.Valid)

	var nt Time
	expect.False(t, nt.Valid)
	nt.Set(time.Now())
	expect.True(t, nt.Valid)

	var nd Date
	expect.False(t, nd.Valid)
	nd.Set(date.From(time.Now()))
	expect.True(t, nd.Valid)

	var nb Bool
	expect.False(t, nb.Valid)
	nb.Set(true)
	expect.True(t, nb.Valid)

	var nu UUID
	expect.False(t, nu.Valid)
	nu.Set(uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
	expect.True(t, nu.Valid)
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
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(bogusString), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(validTrue), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)

	err = json.Unmarshal([]byte(validFalse), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
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
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(rationalFloat), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(roundedFloat), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(validZero), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Int, 0)

	err = json.Unmarshal([]byte(validNegative), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Int, -300)

	err = json.Unmarshal([]byte(validPositive), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Int, 1602525)
}

func TestUnmarshalNullString(t *testing.T) {
	var jsonNull string = `null`
	var jsonNumber string = `3`
	var jsonEmpty string = `""`
	var validString string = `"foo"`

	var n String
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonNumber), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.String, "")

	err = json.Unmarshal([]byte(validString), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.String, "foo")

}

func TestUnmarshalNullTime(t *testing.T) {
	var jsonNull string = `null`
	var jsonEmpty string = `""`
	var stringTime string = `"2010-07-03T13:24:33Z"`
	var stringBogus string = `"bogus"`

	var n Time
	var err error
	err = json.Unmarshal([]byte(jsonNull), &n)
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(stringTime), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Time.Format("2006-01-02 15:04:05"), "2010-07-03 13:24:33")

	err = json.Unmarshal([]byte(stringBogus), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)
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
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &n)
	expect.Nil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(stringDate), &n)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Date.Year, 2010)
	expect.Equal(t, n.Date.Month, time.July)
	expect.Equal(t, n.Date.Day, 3)

	err = json.Unmarshal([]byte(stringTime), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = json.Unmarshal([]byte(stringBogus), &n)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)
}

func TestScanNullTime(t *testing.T) {
	var rawTime = time.Now()
	var mysqlTime = "2010-07-03 13:24:33"
	var byteTime = []byte(mysqlTime)
	var notTime = 3

	var n Time
	var err error
	err = n.Scan(rawTime)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.NotEmpty(t, n.Time.Format("2006-01-02 15:04:05"))

	err = n.Scan(mysqlTime)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Time.Format("2006-01-02 15:04:05"), mysqlTime)

	err = n.Scan(byteTime)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Time.Format("2006-01-02 15:04:05"), mysqlTime)

	err = n.Scan(notTime)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)
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
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Date.Year, 2010)
	expect.Equal(t, n.Date.Month, time.July)
	expect.Equal(t, n.Date.Day, 3)

	err = n.Scan(mysqlTime)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = n.Scan(mysqlDate)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Date.Year, 2010)
	expect.Equal(t, n.Date.Month, time.July)
	expect.Equal(t, n.Date.Day, 3)

	err = n.Scan(byteTime)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)

	err = n.Scan(byteDate)
	expect.Nil(t, err)
	expect.True(t, n.Valid)
	expect.Equal(t, n.Date.Year, 2010)
	expect.Equal(t, n.Date.Month, time.July)
	expect.Equal(t, n.Date.Day, 3)

	err = n.Scan(notTime)
	expect.NotNil(t, err)
	expect.False(t, n.Valid)
}

func TestScanNullUUID(t *testing.T) {
	// start with a null UUID and scan a typical UUID
	{
		expectedUUID := uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
		stringUUID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

		n := UUID{}
		err := n.Scan(stringUUID)
		expect.Nil(t, err, "error unmarshaling null.UUID")
		expect.True(t, n.Valid, "null.UUID should be valid")
		expect.Equal(t, n.UUID, expectedUUID, "UUIDs should be equal")
	}

	// start with some UUID, and scan nil
	{
		n := SomeUUID(uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
		err := n.Scan(nil)
		expect.Nil(t, err, "error unmarshaling null.UUID")
		expect.False(t, n.Valid, "null.UUID should not be valid")
		expect.Equal(t, n.UUID, uuid.Nil, "null.UUID value should be equal to uuid.Nil")
	}
}

func TestValueNullUUID(t *testing.T) {
	u := UUID{}
	val, err := u.Value()
	expect.Nil(t, err, "error getting null.UUID value")
	expect.Nil(t, val, "wrong value returned, should be nil")
}
