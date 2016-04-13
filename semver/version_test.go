package semver

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	examples := []struct {
		String  string
		Version Version
		NotOk   bool
	}{
		{String: "0.0.0", Version: Version{0, 0, 0}},
		{String: "1.0.0", Version: Version{1, 0, 0}},
		{String: "14.54.23", Version: Version{14, 54, 23}},
		{String: "0.2.4", Version: Version{0, 2, 4}},
		{String: "v15.0.3", Version: Version{15, 0, 3}},
		{String: "v9a", Version: Version{9, 0, 0}},
		{String: "v9.1a", Version: Version{9, 1, 0}},

		{String: "hello world", NotOk: true},
		{String: "good 1", NotOk: true},
	}

	for _, example := range examples {
		parsed, ok := Parse(example.String)
		assert.Equal(t, example.Version, parsed, example.String)
		assert.Equal(t, example.NotOk, !ok, example.String)
	}
}

func TestComparisons(t *testing.T) {
	examples := []struct {
		A, B    Version
		Lt, Lte bool
		Gt, Gte bool
	}{
		// TODO: Constraint based testing (ie. https://golang.org/pkg/testing/quick)
		{Version{0, 0, 0}, Version{0, 0, 0}, false, true, false, true},
		{Version{0, 0, 1}, Version{0, 0, 0}, false, false, true, true},
		{Version{0, 1, 0}, Version{0, 0, 0}, false, false, true, true},
		{Version{1, 0, 0}, Version{0, 0, 0}, false, false, true, true},
		{Version{0, 0, 0}, Version{1, 0, 0}, true, true, false, false},
		{Version{0, 0, 1}, Version{1, 0, 0}, true, true, false, false},
		{Version{0, 1, 0}, Version{1, 0, 0}, true, true, false, false},
		{Version{1, 0, 0}, Version{1, 0, 0}, false, true, false, true},

		{Version{1, 2, 3}, Version{3, 2, 1}, true, true, false, false},
		{Version{0, 3, 1}, Version{0, 1, 3}, false, false, true, true},
		{Version{1, 1, 4}, Version{1, 1, 6}, true, true, false, false},
	}

	for _, ex := range examples {
		assert.Equal(t, ex.Lt, ex.A.LessThan(ex.B), ex.A.String()+" < "+ex.B.String())
		assert.Equal(t, ex.Gt, ex.A.GreaterThan(ex.B), ex.A.String()+" > "+ex.B.String())
		assert.Equal(t, ex.Lte, ex.A.AtMost(ex.B), ex.A.String()+" <= "+ex.B.String())
		assert.Equal(t, ex.Gte, ex.A.AtLeast(ex.B), ex.A.String()+" >= "+ex.B.String())
	}
}

func TestString(t *testing.T) {
	examples := []struct {
		Version Version
		String  string
	}{
		{Version: Version{0, 0, 0}, String: "0.0.0"},
		{Version: Version{1, 0, 0}, String: "1.0.0"},
		{Version: Version{14, 54, 23}, String: "14.54.23"},
		{Version: Version{0, 2, 4}, String: "0.2.4"},
		{Version: Version{15, 0, 3}, String: "15.0.3"},
		{Version: Version{9, 0, 0}, String: "9.0.0"},
		{Version: Version{9, 1, 0}, String: "9.1.0"},
	}

	for _, ex := range examples {
		assert.Equal(t, ex.String, ex.Version.String(), ex.String)
	}
}

func TestMarshalJSON(t *testing.T) {
	b1, err1 := json.Marshal(Version{1, 0, 0})
	assert.Nil(t, err1)
	assert.Equal(t, `"1.0.0"`, string(b1))
	b2, err2 := json.Marshal(Version{2, 0, 30})
	assert.Nil(t, err2)
	assert.Equal(t, `"2.0.30"`, string(b2))
}

func TestUnmarshalJSON(t *testing.T) {
	examples := []struct {
		Json    string
		Version Version
	}{
		{`"5.0.0"`, Version{5, 0, 0}},
		{`"v2.4.12"`, Version{2, 4, 12}},
		{`"3.5.0ab"`, Version{3, 5, 0}},
		{`"8.22"`, Version{8, 22, 0}},
	}

	var v Version
	for _, ex := range examples {
		err := json.Unmarshal([]byte(ex.Json), &v)
		assert.Nil(t, err, ex.Json)
		assert.Equal(t, ex.Version, v, ex.Json)
	}

	badExamples := []struct {
		Json  string
		Error string
	}{
		{`null`, "semver: cannot parse version from non-string JSON value"},
		{`""`, "semver: json string is not a valid version"},
		{`"bogus"`, "semver: json string is not a valid version"},
	}

	for _, ex := range badExamples {
		err := json.Unmarshal([]byte(ex.Json), &v)
		if assert.NotNil(t, err, ex.Json) {
			assert.Equal(t, ex.Error, err.Error(), ex.Json)
		}
	}
}
