package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	examples := []struct {
		String  string
		Version Version
		Ok      bool
	}{
		{String: "0.0.0", Version: Version{0, 0, 0}},
		{String: "1.0.0", Version: Version{1, 0, 0}},
		{String: "14.54.23", Version: Version{14, 54, 23}},
		{String: "0.2.4", Version: Version{0, 2, 4}},
		{String: "v15.0.3", Version: Version{15, 0, 3}},
		{String: "v9a", Version: Version{9, 0, 0}},
		{String: "v9.1a", Version: Version{9, 1, 0}},

		{String: "hello world", Ok: false},
		{String: "good 1", Ok: false},
	}

	for _, example := range examples {
		parsed, ok := Parse(example.String)
		assert.Equal(t, example.Version, parsed, example.String)
		assert.Equal(t, example.Ok, ok, example.String)
	}
}

func TestComparisons(t *testing.T) {
	examples := []struct {
		A, B    Version
		Lt, Lte bool
		Gt, Gte bool
	}{
		// Simple combinations
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
