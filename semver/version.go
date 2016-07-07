package semver

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}

var Regexp = regexp.MustCompile(`v?(\d+)(?:\.(\d+))?(?:\.(\d+))?`)
var StrictRegexp = regexp.MustCompile("^" + Regexp.String())

// Parse will parse a semantive version from a string in any of these formats:
//
//     1        // only major
//     1.0      // major/minor
//     1.0.0    // major/minor/patch
//    v1.0.0    // prefixed with "v"
//     1.0.0cc  // with trailing characters (currently ignored)
//
func Parse(input string) (v Version, ok bool) {
	matches := StrictRegexp.FindStringSubmatch(input)
	switch len(matches) {
	case 4:
		v.Patch, _ = strconv.Atoi(matches[3])
		fallthrough
	case 3:
		v.Minor, _ = strconv.Atoi(matches[2])
		fallthrough
	case 2:
		v.Major, _ = strconv.Atoi(matches[1])
		break
	default:
		return
	}

	return v, true
}

func (v Version) LessThan(o Version) bool {
	return v.Major < o.Major ||
		(v.Major == o.Major &&
			(v.Minor < o.Minor ||
				(v.Minor == o.Minor &&
					(v.Patch < o.Patch))))
}

func (v Version) GreaterThan(o Version) bool {
	return v.Major > o.Major ||
		(v.Major == o.Major &&
			(v.Minor > o.Minor ||
				(v.Minor == o.Minor &&
					(v.Patch > o.Patch))))
}

func (v Version) AtLeast(o Version) bool {
	return !o.GreaterThan(v)
}

func (v Version) AtMost(o Version) bool {
	return !o.LessThan(v)
}

// Implements sql.Scanner interface
func (v *Version) Scan(src interface{}) error {
	t, ok := src.([]byte)
	if !ok {
		return errors.New("semver: scan value was not bytes")
	}

	version, ok := Parse(string(t))
	if !ok {
		return errors.New("semver: scan value is not a valid version string")
	}

	*v = version
	return nil
}

// Implements sql.driver.Valuer interface
func (v Version) Value() (driver.Value, error) {
	return v.String(), nil
}

// Implements json.Marshaler interface
func (v Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%v.%v.%v"`, v.Major, v.Minor, v.Patch)), nil
}

// Implements json.Unmarshaler interface
func (v *Version) UnmarshalJSON(bytes []byte) error {
	if bytes[0] != '"' || bytes[len(bytes)-1] != '"' {
		return errors.New("semver: cannot parse version from non-string JSON value")
	}

	parsed, ok := Parse(string(bytes[1 : len(bytes)-1]))
	if !ok {
		return errors.New("semver: json string is not a valid version")
	}

	*v = parsed
	return nil
}
