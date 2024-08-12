package fplib

import "math"

const (
	// Float representation of months in a year.
	mof float64 = 12
	// Float representation of days in a year.
	yrf float64 = 365.25
)

// CalculateMonthlyRate calculates the monthly spending/income rate and returns
// an integer that represents a dollar amount ($100.00 = 10000).
//
// Provide the number of days that the amount has accumulated over, and a rate
// will be returned.
func CalculateMonthlyRate(amount int, days int) int {
	// e.g. 50 days, $100 spent
	//
	// (dollars) / (month) => (dollars) / (365 / 12 days)
	// => (dollars) / (1.64 months) (for 50 days)
	return int(math.Round(float64(amount) / (float64(days) / (yrf / mof))))
}

// CalculateYearlyRate calculates the yearly spending/income rate and returns
// an integer that represents a dollar amount ($100.00 = 10000).
//
// Provide the number of days that the amount has accumulated over, and a rate
// will be returned.
func CalculateYearlyRate(amount int, days int) int {
	// e.g. 400 days, $100 spent
	//
	// (dollars) / (1 year) => (dollars) / (400 / 365 days/year)
	return int(math.Round(float64(amount) / (float64(days) / (yrf))))
}

// CalculateYearlyRate calculates the daily spending/income rate and returns
// an integer that represents a dollar amount ($100.00 = 10000).
//
// Provide the number of days that the amount has accumulated over, and a rate
// will be returned.
func CalculateDailyRate(amount int, days int) int {
	// e.g. 400 days, $100 spent
	//
	// (dollars) / (1 day) => (dollars) / (400 days)
	return int(math.Round(float64(amount) / (float64(days))))
}
