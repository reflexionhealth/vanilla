package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/reflexionhealth/vanilla/date"
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
	marshaler = Int64{}
	assert.NotNil(t, marshaler)
}

func TestImplementsSqlValuer(t *testing.T) {
	var valuer driver.Valuer
	valuer = Date{}
	assert.NotNil(t, valuer)
	valuer = Time{}
	assert.NotNil(t, valuer)
	valuer = String{}
	assert.NotNil(t, valuer)
	valuer = Int64{}
	assert.NotNil(t, valuer)
}

func TestNullDateRefImplementSqlScanner(t *testing.T) {
	var scanner sql.Scanner
	scanner = &Date{}
	assert.NotNil(t, scanner)
	scanner = &Time{}
	assert.NotNil(t, scanner)
	scanner = &String{}
	assert.NotNil(t, scanner)
	scanner = &Int64{}
	assert.NotNil(t, scanner)
}

func TestSetNullable(t *testing.T) {
	var ns String
	assert.False(t, ns.Valid)
	ns.Set("hello world")
	assert.True(t, ns.Valid)

	var ni Int64
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
}

func TestUnmarshalNullTime(t *testing.T) {
	var jsonNull string = `null`
	var jsonEmpty string = `""`
	var stringTime string = `"2010-07-03T13:24:33Z"`
	var stringBogus string = `"bogus"`

	var nt Time
	var err error
	err = json.Unmarshal([]byte(jsonNull), &nt)
	assert.Nil(t, err)
	assert.False(t, nt.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &nt)
	assert.Nil(t, err)
	assert.False(t, nt.Valid)

	err = json.Unmarshal([]byte(stringTime), &nt)
	assert.Nil(t, err)
	assert.True(t, nt.Valid)
	assert.Equal(t, "2010-07-03 13:24:33", nt.Time.Format("2006-01-02 15:04:05"))

	err = json.Unmarshal([]byte(stringBogus), &nt)
	assert.NotNil(t, err)
	assert.False(t, nt.Valid)
}

func TestUnmarshalNullDate(t *testing.T) {
	var jsonNull string = `null`
	var jsonEmpty string = `""`
	var stringDate string = `"2010-07-03"`
	var stringTime string = `"2010-07-03 13:24:33"`
	var stringBogus string = `"bogus"`

	var nd Date
	var err error
	err = json.Unmarshal([]byte(jsonNull), &nd)
	assert.Nil(t, err)
	assert.False(t, nd.Valid)

	err = json.Unmarshal([]byte(jsonEmpty), &nd)
	assert.Nil(t, err)
	assert.False(t, nd.Valid)

	err = json.Unmarshal([]byte(stringDate), &nd)
	assert.Nil(t, err)
	assert.True(t, nd.Valid)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	err = json.Unmarshal([]byte(stringTime), &nd)
	assert.NotNil(t, err)
	assert.False(t, nd.Valid)

	err = json.Unmarshal([]byte(stringBogus), &nd)
	assert.NotNil(t, err)
	assert.False(t, nd.Valid)
}

func TestScanNullTime(t *testing.T) {
	var rawTime = time.Now()
	var mysqlTime = "2010-07-03 13:24:33"
	var byteTime = []byte(mysqlTime)
	var notTime = 3

	var nt Time
	var err error
	err = nt.Scan(rawTime)
	assert.Nil(t, err)
	assert.True(t, nt.Valid)
	assert.NotEmpty(t, nt.Time.Format("2006-01-02 15:04:05"))

	err = nt.Scan(mysqlTime)
	assert.Nil(t, err)
	assert.True(t, nt.Valid)
	assert.Equal(t, mysqlTime, nt.Time.Format("2006-01-02 15:04:05"))

	err = nt.Scan(byteTime)
	assert.Nil(t, err)
	assert.True(t, nt.Valid)
	assert.Equal(t, mysqlTime, nt.Time.Format("2006-01-02 15:04:05"))

	err = nt.Scan(notTime)
	assert.NotNil(t, err)
	assert.False(t, nt.Valid)
}

func TestScanNullDate(t *testing.T) {
	var rawTime = time.Date(2010, time.July, 3, 13, 24, 33, 999, time.UTC)
	var mysqlTime = "2010-07-03 13:24:33"
	var mysqlDate = "2010-07-03"
	var byteTime = []byte(mysqlTime)
	var byteDate = []byte(mysqlDate)
	var notTime = 3

	var nd Date
	var err error
	err = nd.Scan(rawTime)
	assert.Nil(t, err)
	assert.True(t, nd.Valid)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	err = nd.Scan(mysqlTime)
	assert.NotNil(t, err)
	assert.False(t, nd.Valid)

	err = nd.Scan(mysqlDate)
	assert.Nil(t, err)
	assert.True(t, nd.Valid)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	err = nd.Scan(byteTime)
	assert.NotNil(t, err)
	assert.False(t, nd.Valid)

	err = nd.Scan(byteDate)
	assert.Nil(t, err)
	assert.True(t, nd.Valid)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	err = nd.Scan(notTime)
	assert.NotNil(t, err)
	assert.False(t, nd.Valid)
}
