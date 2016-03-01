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
