package fplib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FormatAsCurrency converts an integer to a USD-formatted string. Input
// is assumed to be based in pennies, i.e., hundredths of a dollar - 100 would
// return "$1.00".
func FormatAsCurrency(a int) string {
	if a == 0 {
		return "$0.00"
	}

	if a < 0 {
		s := strconv.Itoa(a * -1)
		if a > -100 {
			return fmt.Sprintf("$-0.%02v", s)
		}

		l := len(s)

		return fmt.Sprintf("$-%v.%v", s[0:l-2], s[l-2:])
	}

	s := strconv.Itoa(a)

	if a < 100 {
		return fmt.Sprintf("$0.%02v", s)
	}

	l := len(s)

	return fmt.Sprintf("$%v.%v", s[0:l-2], s[l-2:])
}

// GetNowDateString returns a string corresponding to the current YYYY-MM-DD
// value.
func GetNowDateString(t time.Time) string {
	return fmt.Sprintf("%04v-%02v-%02v", t.Year(), int(t.Month()), t.Day())
}

// GetDefaultEndDateString returns a string corresponding to the current YYYY-MM-DD
// value plus 1 year in the future, but does not necessarily include 0-padded
// values.
func GetDefaultEndDateString(t time.Time) string {
	return fmt.Sprintf("%04v-%02v-%02v", t.Year()+1, int(t.Month()), t.Day())
}

// GetDateString formats a string as YYYY-MM-DD with zero-padding.
func GetDateString(y, m, d any) string {
	return fmt.Sprintf("%04v-%02v-%02v", y, m, d)
}

// ParseYearMonthDateString takes an input value such as 2020-01-01 and returns
// three integer values - year, month, day. Returns 0, 0, 0 if invalid input
// is received.
func ParseYearMonthDateString(input string) (int, int, int) {
	vals := strings.Split(input, "-")
	if len(vals) != 3 {
		return 0, 0, 0
	}

	yr, _ := strconv.ParseInt(vals[0], 10, 64)
	mo, _ := strconv.ParseInt(vals[1], 10, 64)
	day, _ := strconv.ParseInt(vals[2], 10, 64)

	return int(yr), int(mo), int(day)
}

// This regular expression's purpose is to construct a version of the input
// that only contains digits, periods and nothing else so that it can be
// parsed.
var digitre = regexp.MustCompile(`[^\d.]*`)

// ParseDollarAmount takes an input currency-formatted string, such as $100.00,
// and returns an integer corresponding to the underlying value, such as 10000.
// Generally in this application, values are assumed to be negative (i.e.
// recurring bills), so if assumePositive is set to true, the returned value
// will be positive, but otherwise it will default to negative.
func ParseDollarAmount(input string, assumePositive bool) int64 {
	cents := int64(0)
	multiplier := int64(-1)

	// all values are assumed negative, unless it starts with a + character
	if strings.Index(input, "+") == 0 || strings.Index(input, "$+") == 0 || assumePositive {
		multiplier = int64(1)
	}

	// in the event that the user is entering the starting balance,
	// they may want to set a negative starting balance. So basically just the
	// reverse from above logic, since the user will have to be typing a
	// negative sign in front.
	if assumePositive && (strings.Index(input, "$-") == 0 || strings.Index(input, "-") == 0) {
		multiplier = int64(-1)
	}

	s := digitre.ReplaceAllString(input, "")
	// check if the user entered a period
	ss := strings.Split(s, ".")

	if len(ss) == 2 {
		cents, _ = strconv.ParseInt(ss[1], 10, 64)
		// if the user types e.g. 10.2, they meant $10.20
		// but not if the value started with a 0
		if strings.Index(ss[1], "0") != 0 && cents < 10 {
			cents *= 10
		}
		// if they put in too many numbers, zero it out
		if cents >= 100 {
			cents = 0
		}
	}

	var whole int64
	whole, _ = strconv.ParseInt(ss[0], 10, 64)

	// account for the negative case when re-combining the two values
	// if whole < 0 {
	// 	return multiplier * (whole*100 - cents)
	// }

	return multiplier * (whole*100 + cents)
}

// GetCSVString produces a simple semi-colon-separated value string.
func GetCSVString(input []string) string {
	result := new(strings.Builder)
	if len(input) > 0 {
		result.WriteString(fmt.Sprintf("(%v) ", len(input)))
	}

	for _, name := range input {
		result.WriteString(fmt.Sprintf(`%v; `, name))
	}

	return result.String()
}
