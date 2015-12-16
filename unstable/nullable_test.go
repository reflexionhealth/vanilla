package unstable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
  "testing"

	"github.com/stretchr/testify/assert"
)

func TestImplementsJsonMarshaller(t *testing.T) {
	var marshaler json.Marshaler
	marshaler = NullDate{}
	assert.NotNil(t, marshaler)
	marshaler = NullTime{}
	assert.NotNil(t, marshaler)
	marshaler = NullString{}
	assert.NotNil(t, marshaler)
	marshaler = NullInt64{}
	assert.NotNil(t, marshaler)
}

func TestImplementsSqlValuer(t *testing.T) {
  var valuer driver.Valuer
  valuer = NullDate{}
  assert.NotNil(t, valuer)
  valuer = NullTime{}
  assert.NotNil(t, valuer)
  valuer = NullString{}
  assert.NotNil(t, valuer)
	valuer = NullInt64{}
	assert.NotNil(t, valuer)
}

func TestNullDateRefImplementSqlScanner(t *testing.T) {
	var scanner sql.Scanner
	scanner = &NullDate{}
  assert.NotNil(t, scanner)
	scanner = &NullTime{}
	assert.NotNil(t, scanner)
	scanner = &NullString{}
  assert.NotNil(t, scanner)
	scanner = &NullInt64{}
  assert.NotNil(t, scanner)
}
