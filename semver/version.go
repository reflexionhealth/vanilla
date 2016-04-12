package semver

import (
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
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
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
	var err error
	matches := StrictRegexp.FindStringSubmatch(input)
	switch len(matches) {
	case 4:
		v.Patch, err = strconv.Atoi(matches[3])
		if err != nil {
			ok = false
		}
		fallthrough
	case 3:
		v.Minor, err = strconv.Atoi(matches[2])
		if err != nil {
			ok = false
		}
		fallthrough
	case 2:
		v.Major, _ = strconv.Atoi(matches[1])
		break
	default:
		ok = false
	}
	return
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
