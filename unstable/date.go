package unstable

import (
	"database/sql/driver"
	"errors"
	"time"
)

const (
	MillisecondsInSecond = 1000
	MicrosecondsInSecond = 1000 * MillisecondsInSecond
	NanosecondsInSecond  = 1000 * MicrosecondsInSecond
)

// Date is a plain date, without time or timezone info (use time.Time for those!)
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// Create a Date from a time.Time object
func DateFrom(t time.Time) Date {
	y, m, d := t.Date()
	return Date{y, m, d}
}

func (d Date) PrevDay() Date {
	if d.Day > 1 {
		return Date{d.Year, d.Month, d.Day - 1}
	} else if d.Month > 1 {
		return Date{d.Year, d.Month - 1, DaysInMonth(d.Month-1, d.Year)}
	} else {
		return Date{d.Year - 1, time.December, 31}
	}
}

func (d Date) NextDay() Date {
	if d.Day < DaysInMonth(d.Month, d.Year) {
		return Date{d.Year, d.Month, d.Day + 1}
	} else if d.Month < time.December {
		return Date{d.Year, d.Month + 1, 1}
	} else {
		return Date{d.Year + 1, time.January, 1}
	}
}

func (d Date) Before(other Date) bool {
	return (d.Year < other.Year ||
		(d.Year == other.Year &&
			(d.Month < other.Month ||
				(d.Month == other.Month && d.Day < other.Day))))
}

func (d Date) BeginningOfDay(timezone *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, timezone)
}

func (d Date) EndOfDay(timezone *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 23, 59, 59, NanosecondsInSecond-1, timezone)
}

// Implement sql.Scanner interface
func (d Date) Scan(src interface{}) error {
	t, ok := src.(time.Time)
	if !ok {
		return errors.New("unstable/date: scan value was not a Time")
	}

	d = DateFrom(t)
	return nil
}

// Implement sql.driver.Valuer interface
func (d Date) Value() (driver.Value, error) {
	return d.BeginningOfDay(time.UTC), nil
}

// Implement json.Marshaler interface
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.BeginningOfDay(time.UTC).Format(`"2006-01-02"`)), nil
}

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
