package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Major  int
	Minor  int
	Patch  int
	Build  int
	Commit string
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

var Regexp = regexp.MustCompile(`v?((\d+)(\.\d+)?(\.\d+)?)`)

func Parse(input string) Version {
	var major, minor, patch int
	matches := Regexp.FindStringSubmatch(input)
	switch len(matches) {
	case 5:
		patch, _ = strconv.Atoi(matches[4])
		fallthrough
	case 4:
		minor, _ = strconv.Atoi(matches[3])
		fallthrough
	case 3:
		major, _ = strconv.Atoi(matches[2])
		break
	default:
		major = 1
		break
	}

	return Version{Major: major, Minor: minor, Patch: patch}
}
