package expect

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// True returns true only if the value is true.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.True(t, isSomething)
//    expect.True(t, isSomethingElse, "should be something else")
//
func True(t *testing.T, val interface{}, msg ...interface{}) bool {
	if val != true {
		return errorf(t, "Expected value to be true.", msg...)
	}
	return true
}

// False returns true only if the value is false.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.False(t, isSomething)
//    expect.False(t, isSomethingElse, "should not be something else")
//
func False(t *testing.T, val interface{}, msg ...interface{}) bool {
	if val != false {
		return errorf(t, "Expected value to be false.", msg...)
	}
	return true
}

// Equal returns true only if the two values are the same, are deeply equally,
// or are have different types but equivalent values (eg. int/float).
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.Equal(t, greeting, "Hello world!")
//    expect.Equal(t, 3, 3.0, "value of 3 is not 3")
//
func Equal(t *testing.T, actual, expected interface{}, msg ...interface{}) bool {
	if !areEqual(actual, expected) {
		return errorf(t, fmt.Sprintf("Expected %#v, but got: %#v", expected, actual), msg...)
	}
	return true
}

// NotEqual returns true if the two values are not equivalent.  See Equal.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.NotEqual(t, greeting, "Goodbye planet!")
//    expect.NotEqual(t, 3.0, 3.14, "value of pie is not 3.0")
//
func NotEqual(t *testing.T, actual, expected interface{}, msg ...interface{}) bool {
	if areEqual(actual, expected) {
		return errorf(t, fmt.Sprintf("Expected value not to equal: %#v", expected, actual), msg...)
	}
	return true
}

// areEqual checks if the two values are the same, are deep equally, or are
// different types with equivalent values.
func areEqual(actual, expected interface{}) bool {
	if expected == nil || actual == nil {
		return actual == expected
	}
	actualType := reflect.TypeOf(actual)
	if actualType == reflect.TypeOf(expected) {
		return reflect.DeepEqual(expected, actual)
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(actual, expectedValue.Convert(actualType).Interface())
	}
	return false
}

// Nil returns true only if the value is nil or has an underlying nil value.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.Nil(t, err)
//    expect.Nil(t, err, "err should be noting")
//
func Nil(t *testing.T, val interface{}, msg ...interface{}) bool {
	if !isNil(val) {
		return errorf(t, fmt.Sprintf("Expected nil, but got: %#v", val))
	}
	return true
}

// NotNil returns true only if the value is not nil nor has an underlying nil value.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.NotNil(t, err)
//    expect.NotNil(t, err, "err should be something")
//
func NotNil(t *testing.T, val interface{}, msg ...interface{}) bool {
	if isNil(val) {
		return errorf(t, "Expected value not to be nil.", msg...)
	}
	return true
}

// isNil checks if a value is nil or has an underlying nil value.
func isNil(val interface{}) bool {
	if val == nil {
		return true
	}
	value := reflect.ValueOf(val)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

// Empty returns true only if the value is both empty and has a type
// which can be checked for emptiness (eg. slice, chan, string, map, etc).
// An error is reported with t.Errorf if the expectation is false.
//
// Empty will also check if a value implements Len, Length, Empty, or IsEmpty.
//
//   expect.Empty(t, "")       // true
//   expect.Empty(t, "hello")  // false
//   expect.Empty(t, nil)      // false
//   expect.Empty(t, 0)        // false
//
func Empty(t *testing.T, val interface{}, msg ...interface{}) bool {
	isEmpty, canBeEmpty := isEmpty(val)
	if !canBeEmpty {
		return errorf(t, fmt.Sprintf("Expected value to be empty, but cannot check emptiness of: %v", val), msg...)
	}
	if isEmpty {
		return errorf(t, fmt.Sprintf("Expected value to be empty, but got: %v", val), msg...)
	}
	return true
}

// NotEmpty returns true only if the value is both not empty and has a type
// which can be checked for emptiness (eg. slice, chan, string, map, etc).
// An error is reported with t.Errorf if the expectation is false.
//
// NotEmpty will also check if a value implements Len, Length, Empty, or IsEmpty.
//
//   expect.NotEmpty(t, "hello")  // true
//   expect.NotEmpty(t, "")       // false
//   expect.NotEmpty(t, 13)       // false
//
func NotEmpty(t *testing.T, val interface{}, msg ...interface{}) bool {
	isEmpty, canBeEmpty := isEmpty(val)
	if !canBeEmpty {
		return errorf(t, fmt.Sprintf("Expected value not to be empty, but cannot check emptiness of: %v", val), msg...)
	}
	if isEmpty {
		return errorf(t, "Expected value not to be empty.", msg...)
	}
	return true
}

type implementsLen interface {
	Len() int
}
type implementsLength interface {
	Length() int
}
type implementsEmpty interface {
	Empty() bool
}
type implementsIsEmpty interface {
	IsEmpty() bool
}

// isEmpty checks if a value is a array, chan, map, slice, or string and that
// the value has length zero.
//
// It also considers types with emptiness checking functions.
func isEmpty(val interface{}) (isEmpty bool, canBeEmpty bool) {
	value := reflect.ValueOf(val)
	switch value.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return (value.Len() == 0), true
	}
	switch iv := val.(type) {
	case implementsLen:
		return (iv.Len() == 0), true
	case implementsLength:
		return (iv.Length() == 0), true
	case implementsEmpty:
		return iv.Empty(), true
	case implementsIsEmpty:
		return iv.IsEmpty(), true
	}
	return false, false
}

// Contains returns true only if the value is a string, array, slice, or map and
// the value contains the specified substring, element, or key.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.Contains(t, [3, 62, 11], 11)
//    expect.Contains(t, "Hello world", "world")
//    expect.Contains(t, {"key": "value"}, "key")
//
func Contains(t *testing.T, set, elem interface{}, msg ...interface{}) bool {
	hasElement, isContainer := containsElement(set, elem)
	if !isContainer {
		return errorf(t, fmt.Sprintf("Expected value to be a container, but got: %v", set), msg...)
	}
	if !hasElement {
		return errorf(t, fmt.Sprintf("Expected \"%s\" to contain \"%s\"", set, elem), msg...)
	}
	return true
}

// NotContains returns true only if the value is a string, array, slice, or map
// and the value does not contain the specified substring, element, or key.
// An error is reported with t.Errorf if the expectation is false.
//
//    expect.NotContains(t, [3, 62, 11], 24)
//    expect.NotContains(t, "Hello world", "earth")
//    expect.NotContains(t, {"a": "apple"}, "b")
//
func NotContains(t *testing.T, set, elem interface{}, msg ...interface{}) bool {
	hasElement, isContainer := containsElement(set, elem)
	if !isContainer {
		return errorf(t, fmt.Sprintf("Expected value to be a container, but got: %v", set), msg...)
	}
	if hasElement {
		return errorf(t, fmt.Sprintf("Expected \"%s\" not to contain \"%s\"", set, elem), msg...)
	}
	return true
}

// containsElement checks if a value is a string, array, slice, or map and
// whether that container contains a specified element.
func containsElement(set interface{}, elem interface{}) (hasElement, isContainer bool) {
	setKind := reflect.TypeOf(set).Kind()
	setVal := reflect.ValueOf(set)
	elemVal := reflect.ValueOf(elem)
	if setKind == reflect.String {
		if reflect.TypeOf(elem).Kind() == reflect.String {
			return strings.Contains(setVal.String(), elemVal.String()), true
		}
		return false, true
	}
	if setKind == reflect.Map {
		keys := setVal.MapKeys()
		for _, key := range keys {
			if key == reflect.ValueOf(elem) {
				return true, true
			}
		}
		return false, true
	}
	if setKind == reflect.Array || setKind == reflect.Slice {
		for i := 0; i < setVal.Len(); i++ {
			if setVal.Index(i).Interface() == elem {
				return true, true
			}
		}
		return false, true
	}
	return false, false
}

// AlmostEqual returns true if the actual and expected numerals are within the
// specified delta of each other.
// An error is reported with t.Errorf if the expectation is false.
//
// 	 expect.AlmostEqual(t, math.Pi, (22 / 7.0))
// 	 expect.AlmostEqual(t, math.Pi, (22 / 7.0), 0.05)
//
func AlmostEqual(t *testing.T, actual, expected interface{}, deltaOrMsg ...interface{}) bool {
	var msg []interface{}
	var delta float64
	if len(deltaOrMsg) > 0 {
		if d, ok := toFloat(deltaOrMsg[0]); ok {
			msg = deltaOrMsg[1:]
			delta = d
		}
	}

	a, ok := toFloat(actual)
	if !ok {
		return errorf(t, fmt.Sprintf("Expected a number, but got: %v", actual), msg...)
	} else if math.IsNaN(a) {
		return errorf(t, fmt.Sprintf("Expected a number, but got: NaN"), msg...)
	}

	b, ok := toFloat(expected)
	if !ok {
		return errorf(t, fmt.Sprintf("Expected a number, but got: %v", expected), msg...)
	}

	dt := a - b
	if dt < -delta || dt > delta {
		return errorf(t, fmt.Sprintf("Expected %v to be within %v of %v, but difference was %v", actual, expected, delta, dt), msg...)
	}

	return true
}

var floatType = reflect.TypeOf(0.0)

func toFloat(val interface{}) (float64, bool) {
	if reflect.TypeOf(val).ConvertibleTo(floatType) {
		return reflect.ValueOf(val).Convert(floatType).Float(), true
	}
	return 0.0, false
}

// Regexp returns true if the value formatted as a string matches the regexp.
// An error is reported with t.Errorf if the expectation is false.
//
//  expect.Regexp(t, "it's starting", regexp.MustCompile("start"))
//  expect.Regexp(t, "it's not starting", "start...$")
//
func Regexp(t *testing.T, str interface{}, exp interface{}, msg ...interface{}) bool {
	if !matchRegexp(exp, str) {
		return errorf(t, fmt.Sprintf("Expected \"%v\" to match \"%v\"", str, exp), msg...)
	}
	return true
}

// NotRegexp returns true if the value formatted as a string does not match the regexp.
// An error is reported with t.Errorf if the expectation is false.
//
//  expect.NotRegexp(t, "it's starting", regexp.MustCompile("starts"))
//  expect.NotRegexp(t, "it's not starting", "start$")
//
func NotRegexp(t *testing.T, str interface{}, exp interface{}, msg ...interface{}) bool {
	if matchRegexp(exp, str) {
		return errorf(t, fmt.Sprintf("Expected \"%v\" to NOT match \"%v\"", str, exp), msg...)
	}
	return true
}

// matchRegexp checks if a value is a string and matches a specified regexp.
func matchRegexp(exp interface{}, str interface{}) bool {
	var r *regexp.Regexp
	if rr, ok := exp.(*regexp.Regexp); ok {
		r = rr
	} else {
		r = regexp.MustCompile(fmt.Sprint(exp))
	}
	return (r.FindStringIndex(fmt.Sprint(str)) != nil)

}

// errorf emits an error message for a failed assertion and always returns false.
func errorf(t *testing.T, expectation string, msg ...interface{}) bool {
	stacktrace := strings.Join(getStacktrace(), "\n\r\t\t ")
	if len(msg) > 0 {
		t.Errorf("\r%s\r\tMessage: %s\n\r\t  Error: %s\n\r\t  Trace: %s\n\r",
			getWhitespaceString(),
			fmt.Sprintf(msg[0].(string), msg[1:]...),
			expectation,
			stacktrace)
	} else {
		t.Errorf("\r%s\r\t  Error: %s\n\r\t  Trace: %s\n\r",
			getWhitespaceString(),
			expectation,
			stacktrace)
	}
	return false
}

// NOTE: Mostly stolen from "github.com/stretchr/testify".
// getStacktrace return the current stacktrace, ignoring frames in this package.
func getStacktrace() []string {
	pc := uintptr(0)
	file := ""
	line := 0
	ok := false
	name := ""

	callers := []string{}
	for i := 0; ; i++ {
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			return nil
		}

		if file == "<autogenerated>" {
			break
		}

		if strings.HasSuffix(file, "/expect/expect.go") {
			continue
		}

		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
		callers = append(callers, fmt.Sprintf("%s:%d", file, line))
		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}

		name = f.Name()
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}

	return callers
}

// NOTE: Stolen from the `go test` tool.
// isTest tells whether name looks like a test (or benchmark, according to prefix).
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer.
func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
	}
	rune, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(rune)
}

// NOTE: Mysticism stolen from "github.com/stretchr/testify".
// getWhitespaceString returns a string that is long enough to overwrite the
// default output from the go testing framework.
func getWhitespaceString() string {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	return strings.Repeat(" ", len(fmt.Sprintf("%s:%d:        ", file, line)))
}
