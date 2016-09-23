package semver

import (
	"encoding/json"
	"testing"

	"github.com/reflexionhealth/vanilla/expect"
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
		expect.Equal(t, parsed, example.Version, example.String)
		expect.Equal(t, !ok, example.NotOk, example.String)
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
		expect.Equal(t, ex.A.LessThan(ex.B), ex.Lt, ex.A.String()+" < "+ex.B.String())
		expect.Equal(t, ex.A.GreaterThan(ex.B), ex.Gt, ex.A.String()+" > "+ex.B.String())
		expect.Equal(t, ex.A.AtMost(ex.B), ex.Lte, ex.A.String()+" <= "+ex.B.String())
		expect.Equal(t, ex.A.AtLeast(ex.B), ex.Gte, ex.A.String()+" >= "+ex.B.String())
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
		expect.Equal(t, ex.Version.String(), ex.String)
	}
}

func TestMarshalJSON(t *testing.T) {
	b1, err1 := json.Marshal(Version{1, 0, 0})
	expect.Nil(t, err1)
	expect.Equal(t, string(b1), `"1.0.0"`)
	b2, err2 := json.Marshal(Version{2, 0, 30})
	expect.Nil(t, err2)
	expect.Equal(t, string(b2), `"2.0.30"`)
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
		expect.Nil(t, err, ex.Json)
		expect.Equal(t, v, ex.Version, ex.Json)
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
		if expect.NotNil(t, err, ex.Json) {
			expect.Equal(t, err.Error(), ex.Error, ex.Json)
		}
	}
}
