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

func TestScanNullTime(t *testing.T) {
	var rawTime = time.Now()
	var stringTime string = "2010-07-03 13:24:33"
	var byteTime = []byte(stringTime)
	var notTime = 3

	var nt Time
	nt.Scan(rawTime)
	assert.True(t, nt.Valid)
	assert.NotNil(t, nt.Time)

	nt.Scan(stringTime)
	assert.True(t, nt.Valid)
	assert.NotNil(t, nt.Time)
	assert.Equal(t, stringTime, nt.Time.Format("2006-01-02 15:04:05"))

	nt.Scan(byteTime)
	assert.True(t, nt.Valid)
	assert.NotNil(t, nt.Time)
	assert.Equal(t, stringTime, nt.Time.Format("2006-01-02 15:04:05"))

	err := nt.Scan(notTime)
	assert.NotNil(t, err)
}

func TestScanNullDate(t *testing.T) {
	var rawTime = time.Date(2010, time.July, 3, 13, 24, 33, 999, time.UTC)
	var stringTime string = "2010-07-03 13:24:33"
	var stringDate string = "2010-07-03"
	var byteTime = []byte(stringTime)
	var byteDate = []byte(stringDate)
	var notTime = 3

	var nd Date
	nd.Scan(rawTime)
	assert.True(t, nd.Valid)
	assert.NotNil(t, nd.Date)

	nd.Scan(stringTime)
	assert.True(t, nd.Valid)
	assert.NotNil(t, nd.Date)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	nd.Scan(stringDate)
	assert.True(t, nd.Valid)
	assert.NotNil(t, nd.Date)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	nd.Scan(byteTime)
	assert.True(t, nd.Valid)
	assert.NotNil(t, nd.Date)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	nd.Scan(byteDate)
	assert.True(t, nd.Valid)
	assert.NotNil(t, nd.Date)
	assert.Equal(t, 2010, nd.Date.Year)
	assert.Equal(t, time.July, nd.Date.Month)
	assert.Equal(t, 3, nd.Date.Day)

	err := nd.Scan(notTime)
	assert.NotNil(t, err)
}
