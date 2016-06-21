package date

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/reflexionhealth/vanilla/clock"
)

var Clock *clock.Source = &clock.Default

const (
	MillisecondsInSecond = 1000
	MicrosecondsInSecond = 1000 * MillisecondsInSecond
	NanosecondsInSecond  = 1000 * MicrosecondsInSecond
)

const RFC3339 = "2006-01-02"

// Parse string into desired date format i.e RFC3339
func Parse(format string, source string) (Date, error) {
	t, err := time.Parse(format, source)
	return From(t), err
}

// Date is a plain date, without time or timezone info (use time.Time for those!)
type Date struct {
	Year  int
	Month time.Month
	Day   int

	// NOTE: time.Time does not preserve timezone when gob encoded
	// location must be private to support Gob encoding
	location *time.Location
}

func At(y int, m time.Month, d int, l *time.Location) Date {
	return Date{y, m, d, l}
}

func TodayIn(loc *time.Location) Date {
	t := Clock.In(loc)
	y, m, d := t.Date()
	return Date{y, m, d, loc}
}

func TodayUtc() Date {
	return TodayIn(time.UTC)
}

func YesterdayIn(loc *time.Location) Date {
	return TodayIn(loc).PrevDay()
}

// Create a Date from a time.Time object
func From(t time.Time) Date {
	y, m, d := t.Date()
	return Date{y, m, d, t.Location()}
}

func (d Date) DaysAfter(other Date) int {
	return int(d.BeginningOfDayIn(time.UTC).Sub(other.BeginningOfDayIn(time.UTC)).Hours() / 24)
}

func (d Date) AddDays(num int) Date {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, d.location).AddDate(0, 0, num)
	year, month, day := t.Date()
	return Date{year, month, day, d.location}
}

func (d Date) PrevDay() Date {
	if d.Day > 1 {
		return Date{d.Year, d.Month, d.Day - 1, d.location}
	} else if d.Month > 1 {
		return Date{d.Year, d.Month - 1, DaysInMonth(d.Month-1, d.Year), d.location}
	} else {
		return Date{d.Year - 1, time.December, 31, d.location}
	}
}

func (d Date) NextDay() Date {
	if d.Day < DaysInMonth(d.Month, d.Year) {
		return Date{d.Year, d.Month, d.Day + 1, d.location}
	} else if d.Month < time.December {
		return Date{d.Year, d.Month + 1, 1, d.location}
	} else {
		return Date{d.Year + 1, time.January, 1, d.location}
	}
}

// MostRecent returns today if today is the given weekday, otherwise it returns
// the date of the last day for that weekday.
//
// It is useful to get the beginning of the week:
//
//    today.MostRecent(time.Sunday) // in the US
//    today.MostRecent(time.Monday) // officially
//
func (d Date) MostRecent(weekday time.Weekday) Date {
	t := d.BeginningOfDay()
	today := t.Weekday()
	if today == weekday {
		return d
	} else if today > weekday {
		return From(t.AddDate(0, 0, int(weekday-today)))
	} else {
		return From(t.AddDate(0, 0, int(weekday-today)-7))
	}
}

// MostSoon returns today if today is the given weekday, otherwise it returns
// the date of the upcoming day for that weekday.
func (d Date) MostSoon(weekday time.Weekday) Date {
	if d.Weekday() == weekday {
		return d
	} else if d.Weekday() < weekday {
		return d.AddDays(int(weekday - d.Weekday()))
	} else {
		return d.AddDays(7 - int(d.Weekday()-weekday))
	}
}

func (d Date) Before(other Date) bool {
	return (d.Year < other.Year ||
		(d.Year == other.Year &&
			(d.Month < other.Month ||
				(d.Month == other.Month && d.Day < other.Day))))
}

func (d Date) Equal(other Date) bool {
	return d == other
}

func (d Date) After(other Date) bool {
	return (d.Year > other.Year ||
		(d.Year == other.Year &&
			(d.Month > other.Month ||
				(d.Month == other.Month && d.Day > other.Day))))
}

func (d Date) AtLeast(other Date) bool {
	return !other.After(d)
}

func (d Date) AtMost(other Date) bool {
	return !other.Before(d)
}

func (d Date) Weekday() time.Weekday {
	return d.BeginningOfDay().Weekday()
}

func (d Date) BeginningOfDay() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, d.location)
}

func (d Date) BeginningOfDayIn(timezone *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, timezone)
}

func (d Date) EndOfDay() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 23, 59, 59, NanosecondsInSecond-1, d.location)
}

func (d Date) EndOfDayIn(timezone *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 23, 59, 59, NanosecondsInSecond-1, timezone)
}

func (d Date) String() string {
	return d.Format("2006-01-02")
}

func (d Date) Format(format string) string {
	return d.BeginningOfDayIn(time.UTC).Format(format)
}

// Implements sql.Scanner interface
func (d *Date) Scan(src interface{}) error {
	t, ok := src.(time.Time)
	if !ok {
		return errors.New("date: scan value was not a Time")
	}

	*d = From(t)
	return nil
}

// Implements sql.driver.Valuer interface
func (d Date) Value() (driver.Value, error) {
	return d.BeginningOfDayIn(time.UTC), nil
}

// Implements json.Marshaler interface
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Format(`"2006-01-02"`)), nil
}

// Implements json.Unmarshaler interface
func (d *Date) UnmarshalJSON(bytes []byte) error {
	t, err := time.Parse(`"2006-01-02"`, string(bytes))
	if err != nil {
		return err
	}

	*d = From(t)
	return nil
}

// TODO: Implment gob.GobEncoder and gob.GobDecoder to preserve timezone
// func (d Date) GobEncode() ([]byte, error) {}
// func (d *Date) GobDecode(bytes []byte) error {}

func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

var DaysInNonLeapMonth = [12]int32{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func DaysInMonth(m time.Month, year int) int {
	if m == time.February && IsLeapYear(year) {
		return 29
	} else {
		return int(DaysInNonLeapMonth[m-1])
	}
}
