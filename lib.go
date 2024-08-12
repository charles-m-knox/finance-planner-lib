package fplib

import (
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"git.cmcode.dev/cmcode/uuid"
	"github.com/teambition/rrule-go"
)

// rrule.ROption
// var examplerro rrule.ROption = rrule.ROption{
// 	Freq:     rrule.DAILY, // occurs daily
// 	Interval: 1,           // occurs every X days
// 	Count:    5,           // ends after 5 occurrences
// 	Dtstart:  time.Now(),
// 	Until:    time.Now(),

// 	// 1: First occurrence of the weekday within the month
// 	// 2: Second occurrence of the weekday within the month
// 	// -1: Last occurrence of the weekday within the month
// 	// -2: Second-to-last occurrence of the weekday within the month
// 	Bysetpos: []int{1, 2, -1, -2},
// }

type TX struct { // transaction
	// Order  int    `yaml:"order"`  // manual ordering
	Amount int    `yaml:"amount"` // in cents; 500 = $5.00
	Active bool   `yaml:"active"`
	Name   string `yaml:"name"`
	Note   string `yaml:"note"`
	// for examples of rrules:
	// https://github.com/teambition/rrule-go/blob/f71921a2b0a18e6e73c74dea155f3a549d71006d/rrule.go#L91
	// https://github.com/teambition/rrule-go/blob/master/rruleset_test.go
	// https://labix.org/python-dateutil/#head-88ab2bc809145fcf75c074817911575616ce7caf
	RRule string `yaml:"rrule"`

	// for when users don't want to use the rrules:

	// The frequency of recurrence, such as MONTHLY/YEARLY/DAILY.
	Frequency string `yaml:"frequency"`
	// The interval of recurrence. A value of 1 means that this occurs every
	// 1 month/year/day. A value of 6 means that this occurs every 6th month/year/day.
	Interval    int          `yaml:"interval"`
	Weekdays    map[int]bool `yaml:"weekdays"` // monday starts on 0
	StartsDay   int          `yaml:"startsDay"`
	StartsMonth int          `yaml:"startsMonth"`
	StartsYear  int          `yaml:"startsYear"`
	EndsDay     int          `yaml:"endsDay"`
	EndsMonth   int          `yaml:"endsMonth"`
	EndsYear    int          `yaml:"endsYear"`
	ID          string       `yaml:"id"`
	CreatedAt   time.Time    `yaml:"createdAt"`
	UpdatedAt   time.Time    `yaml:"updatedAt"`
	Selected    bool         `yaml:"selected"` // when activated in the transactions table
}

type PreCalculatedResult struct {
	Date                  time.Time
	DayTransactionNames   []string
	DayTransactionAmounts []int
}

// A result is a csv/table output row as shown in a results page.
type Result struct {
	Record                   int
	Date                     time.Time
	Balance                  int
	CumulativeIncome         int
	CumulativeExpenses       int
	DayExpenses              int
	DayIncome                int
	DayNet                   int
	DayTransactionNames      string
	DiffFromStart            int
	DayTransactionNamesSlice []string
	ID                       string
	CreatedAt                string
	UpdatedAt                string
}

// GetNewTX returns an empty transaction with sensible defaults based on the
// provided time t (which is typically time.Now()).
func GetNewTX(t time.Time) TX {
	oneMonth := t.Add(time.Hour * HoursInDay * DaysInMonth)

	return TX{
		// Order:       0,
		Amount:      DefaultTransactionBalance,
		Active:      true,
		Name:        New,
		Frequency:   MONTHLY,
		Interval:    1,
		StartsDay:   t.Day(),
		StartsMonth: int(t.Month()),
		StartsYear:  t.Year(),
		EndsDay:     oneMonth.Day(),
		EndsMonth:   int(oneMonth.Month()),
		EndsYear:    oneMonth.Year(),
		ID:          uuid.New(),
		CreatedAt:   t,
		UpdatedAt:   t,
		Note:        "",
		RRule:       "",
		Weekdays:    GetWeekdaysMap(),
		Selected:    false,
	}
}

// GetWeekdaysMap returns a map that can be used like this:
//
// m := GetWeekdaysMap()
//
// if m[rrule.MO.Day()] { /* do something * / }
//
// It is meant to be more efficient than repeatedly using tx.HasWeekday()
// to determine if a weekday is present in a given TX.
// func (tx *TX) GetWeekdaysMap() map[int]bool {
// 	m := make(map[int]bool)
// 	for i := 0; i < 7; i++ {
// 		m[i] = false
// 	}

// 	for i := range tx.Weekdays {
// 		m[tx.Weekdays[i]] = true
// 	}

// 	return m
// }

// GetEmptyWeekdaysMap returns a map that can be used like this:
//
// m := GetWeekdaysMap()
//
// if m[rrule.MO.Day()] { /* do something * / }
//
// It is meant to be more efficient than repeatedly using tx.HasWeekday()
// to determine if a weekday is present in a given TX.
func GetWeekdaysMap() map[int]bool {
	m := make(map[int]bool)
	for i := 0; i < 7; i++ {
		m[i] = false
	}

	return m
}

// GetWeekdaysCheckedMap returns a map that can be used like this:
//
// checkedGlyph := "X"
// uncheckedGlyph := " "
// m := GetWeekdaysCheckedMap(checkedGlyph)
//
// log.Println("occurs on mondays: %v", m[rrule.MO.Day()])
//
// It is meant to be more efficient than repeatedly using tx.HasWeekday()
// to determine if a weekday is present in a given TX.
func (tx *TX) GetWeekdaysCheckedMap(checked, unchecked string) map[int]string {
	m := make(map[int]string)

	for k, v := range tx.Weekdays {
		if !v {
			m[k] = unchecked

			continue
		}

		m[k] = checked
	}

	return m
}

// HasWeekday checks if a recurring transaction definition contains
// the specified weekday as an rrule recurrence day of the week.
// func (tx *TX) HasWeekday(weekday int) bool {
// 	for k, v := range tx.Weekdays {
// 		if weekday == d {
// 			return true
// 		}
// 	}

// 	return false
// }

// func ToggleDayFromWeekdays(weekdays []int, weekday int) []int {
// 	if weekday < 0 || weekday > 6 {
// 		return weekdays
// 	}

// 	foundWeekday := false
// 	returnValue := []int{}

// 	for i := range weekdays {
// 		if weekdays[i] == weekday {
// 			foundWeekday = true
// 		} else {
// 			returnValue = append(returnValue, weekdays[i])
// 		}
// 	}

// 	if !foundWeekday {
// 		returnValue = append(returnValue, weekday)
// 	}

// 	sort.Ints(returnValue)

// 	return returnValue
// }

func GetResults(tx []TX, startDate time.Time, endDate time.Time, startBalance int, statusHook func(status string)) ([]Result, error) {
	if startDate.After(endDate) {
		return []Result{}, fmt.Errorf("start date is after end date: %v vs %v", startDate, endDate)
	}

	// start by quickly generating an index of every single date from startDate to endDate
	dates := make(map[int64]Result)
	preCalculatedDates := make(map[int64]PreCalculatedResult)

	r, err := rrule.NewRRule(
		rrule.ROption{
			Freq:    rrule.DAILY,
			Dtstart: startDate,
			Until:   endDate,
		},
	)
	if err != nil {
		return []Result{}, fmt.Errorf("failed to construct rrule for results date window: %v", err.Error())
	}

	allDates := r.All()

	statusHook("preparing dates...")

	for i, dt := range allDates {
		dtInt := dt.Unix()
		dates[dtInt] = Result{
			Record: i,
			Date:   dt,
		}
		preCalculatedDates[dtInt] = PreCalculatedResult{
			Date: dt,
		}
	}

	emptyDate := time.Date(0, time.Month(0), 0, 0, 0, 0, 0, time.UTC)

	// iterate over every TX definition, starting with its start date
	txLen := len(tx)

	statusHook(fmt.Sprintf("recurrences... [%v/%v]", 0, txLen))

	for i, txi := range tx {
		if !txi.Active {
			continue
		}

		if i%1000 == 0 {
			// to avoid unnecessary slowdown, only update every 1000 iterations
			statusHook(fmt.Sprintf("recurrences... [%v/%v]", i+1, txLen))
		}

		var allOccurrences []time.Time

		if txi.RRule != "" {
			s, err := rrule.StrToRRuleSet(txi.RRule)
			if err != nil {
				return []Result{}, fmt.Errorf(
					"failed to process rrule for tx %v: %v",
					txi.Name,
					err.Error(),
				)
			}

			allOccurrences = s.Between(
				startDate,
				endDate,
				true,
			)
		} else {
			txiStartsDate := time.Date(txi.StartsYear, time.Month(txi.StartsMonth), txi.StartsDay, 0, 0, 0, 0, time.UTC)
			txiEndsDate := time.Date(txi.EndsYear, time.Month(txi.EndsMonth), txi.EndsDay, 0, 0, 0, 0, time.UTC)
			// input validation: if the end date for the transaction definition is after
			// the final end date, then just use the ending date.
			// also, if the transaction definition's end date is unset (equal to emptyDate),
			// then default to the ending date as well
			if txiEndsDate.After(endDate) || txiEndsDate == emptyDate {
				txiEndsDate = endDate
			}
			// input validation: if the transaction definition's start date is
			// unset (equal to emptyDate), then default to the start date
			if txiStartsDate == emptyDate {
				txiStartsDate = startDate
			}

			// This is the rrule that determines the recurrence pattern for
			// this particular transaction definition.
			var s *rrule.RRule
			// These are the rrule options that we are construction for this
			// particular transaction definition.
			var rr rrule.ROption

			// convert the user-input weekdays into a value that rrule lib will
			// accept
			weekdays := []rrule.Weekday{}

			for wi, active := range txi.Weekdays {
				if !active {
					continue
				}

				switch wi {
				case rrule.MO.Day():
					weekdays = append(weekdays, rrule.MO)
				case rrule.TU.Day():
					weekdays = append(weekdays, rrule.TU)
				case rrule.WE.Day():
					weekdays = append(weekdays, rrule.WE)
				case rrule.TH.Day():
					weekdays = append(weekdays, rrule.TH)
				case rrule.FR.Day():
					weekdays = append(weekdays, rrule.FR)
				case rrule.SA.Day():
					weekdays = append(weekdays, rrule.SA)
				case rrule.SU.Day():
					weekdays = append(weekdays, rrule.SU)
				default:
					break
				}
			}

			rr.Dtstart = txiStartsDate
			rr.Until = txiEndsDate
			rr.Interval = txi.Interval

			// TODO: this code is unable to support weekdays when using
			// yearly/monthly recurrence patterns. This library needs to
			// increment to the next version and the UIs need to be updated
			// to support this capability.

			switch txi.Frequency {
			case rrule.YEARLY.String():
				rr.Freq = rrule.YEARLY
			case rrule.MONTHLY.String():
				rr.Freq = rrule.MONTHLY
			default:
				rr.Freq = rrule.DAILY
				rr.Byweekday = weekdays
			}

			s, err = rrule.NewRRule(rr)
			if err != nil {
				return []Result{}, fmt.Errorf(
					"failed to construct rrule for tx %v: %v",
					txi.Name,
					err.Error(),
				)
			}

			allOccurrences = s.Between(startDate, endDate, true)
		}

		for _, dt := range allOccurrences {
			dtInt := dt.Unix()
			newResult := preCalculatedDates[dtInt]
			newResult.Date = dt
			newResult.DayTransactionAmounts = append(newResult.DayTransactionAmounts, txi.Amount)
			newResult.DayTransactionNames = append(newResult.DayTransactionNames, txi.Name)
			preCalculatedDates[dtInt] = newResult
		}
	}

	results := []Result{}
	for _, result := range dates {
		results = append(results, result)
	}

	resultsLen := len(results)
	statusHook(fmt.Sprintf("sorting dates... [%v]", resultsLen))
	sort.SliceStable(
		results,
		func(i, j int) bool {
			return results[j].Date.After(results[i].Date)
		},
	)

	// now that it's sorted, we can roll out the calculations
	currentBalance := startBalance
	diff := 0
	cumulativeIncome := 0
	cumulativeExpenses := 0

	statusHook(fmt.Sprintf("calculating... [%v/%v]", 0, resultsLen))

	for i := range results {
		if i%1000 == 0 {
			// to avoid unnecessary slowdown, only update every 1000 iterations
			statusHook(fmt.Sprintf("calculating... [%v/%v]", i+1, resultsLen))
		}

		resultsDateInt := results[i].Date.Unix()
		numDayTransactionAmounts := len(preCalculatedDates[resultsDateInt].DayTransactionAmounts)
		numDdayTransactionNames := len(preCalculatedDates[resultsDateInt].DayTransactionNames)

		// if for some reason not all transaction names and amounts match up,
		// exit now
		if numDayTransactionAmounts != numDdayTransactionNames {
			return results, fmt.Errorf(
				"there was a different number of transaction amounts versus transaction names for date %v",
				resultsDateInt,
			)
		}

		for j := range preCalculatedDates[resultsDateInt].DayTransactionAmounts {
			// determine if the amount is an expense or income
			amt := preCalculatedDates[resultsDateInt].DayTransactionAmounts[j]
			if amt >= 0 {
				results[i].DayIncome += amt
				cumulativeIncome += amt
			} else {
				results[i].DayExpenses += amt
				cumulativeExpenses += amt
			}

			// basically just doing a join on a slice of strings, should
			// use the proper method for this in the future
			name := preCalculatedDates[resultsDateInt].DayTransactionNames[j]
			if results[i].DayTransactionNames == "" {
				results[i].DayTransactionNames = name
			} else {
				results[i].DayTransactionNames += fmt.Sprintf("; %v", name)
			}

			results[i].DayTransactionNamesSlice = append(results[i].DayTransactionNamesSlice, name)

			results[i].DayNet += amt
			diff += amt
			currentBalance += amt
		}

		results[i].Balance = currentBalance
		results[i].CumulativeIncome = cumulativeIncome
		results[i].CumulativeExpenses = cumulativeExpenses
		results[i].DiffFromStart = diff
	}

	statusHook(fmt.Sprintf("done [%v/%v]", resultsLen, resultsLen))

	return results, nil
}

// GetStartDateString returns a formatted date string for the transaction's
// start date.
func (tx *TX) GetStartDateString() string {
	return GetDateString(tx.StartsYear, tx.StartsMonth, tx.StartsDay)
}

// GetEndsDateString returns a formatted date string for the transaction's end
// date.
func (tx *TX) GetEndsDateString() string {
	return GetDateString(tx.EndsYear, tx.EndsMonth, tx.EndsDay)
}

// RemoveTXAtIndex is a quick helper function to remove a transaction from
// a slice. There are more generic ways to do this, and it's fairly trivial,
// but it's nice to have a dedicated helper function for it.
func RemoveTXAtIndex(txs []TX, i int) []TX {
	return append(txs[:i], txs[i+1:]...)
}

// RemoveTXByID manipulates an input TX slice by removing a TX with the provided
// id.
func RemoveTXByID(txs *[]TX, id string) {
	for i := range *txs {
		tx := (*txs)[i]

		if tx.ID != id {
			continue
		}

		*txs = RemoveTXAtIndex(*txs, i)

		break
	}
}

// GetTXByID finds the index of a TX for the provided id, returning an error
// and -1 if not present.
func GetTXByID(txs *[]TX, id string) (int, error) {
	for i := range *txs {
		tx := (*txs)[i]

		if tx.ID != id {
			continue
		}

		return i, nil
	}

	return -1, errors.New("not present")
}

// GetDateFromStrSafe converts a provided string, typically formatted like
// "YYYY-MM-DD" into a valid time.Time. If the provided string is not formatted
// in this manner, then the argument t is used as the resulting time.Time value
// instead, with its date being set to t.Year, t.Month, t.Day and its
// hours/minutes/seconds being set to zero.
func GetDateFromStrSafe(s string, t time.Time) time.Time {
	if s == "0-0-0" || s == "--" || s == "" {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	y, m, d := ParseYearMonthDateString(s)

	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// GenerateResultsFromDateStrings takes an input start and end date (either can
// be the default '0-0-0' values, in which case it uses today for the start,
// and a year from now for the end), and calculates all of the calculable
// transactions for the provided range.
// func GenerateResultsFromDateStrings(
// 	txs *[]TX,
// 	bal int,
// 	startDt string,
// 	endDt string,
// 	statusHook func(status string),
// ) ([]Result, error) {
// 	now := time.Now()
// 	stYr, stMo, stDay := ParseYearMonthDateString(startDt)
// 	endYr, endMo, endDay := ParseYearMonthDateString(endDt)

// 	if startDt == "0-0-0" || startDt == "--" || startDt == "" {
// 		stYr = now.Year()
// 		stMo = int(now.Month())
// 		stDay = now.Day()
// 	}

// 	if endDt == "0-0-0" || endDt == "--" || endDt == "" {
// 		endYr = now.Year() + 1
// 		endMo = int(now.Month())
// 		endDay = now.Day()
// 	}

// 	res, err := GetResults(
// 		*txs,
// 		time.Date(stYr, time.Month(stMo), stDay, 0, 0, 0, 0, time.UTC),
// 		time.Date(endYr, time.Month(endMo), endDay, 0, 0, 0, 0, time.UTC),
// 		bal,
// 		statusHook,
// 	)
// 	if err != nil {
// 		return []Result{}, fmt.Errorf("failed to get results: %v", err.Error())
// 	}

// 	return res, nil
// }

type TXStats struct {
	DailySpending   int
	DailyIncome     int
	DailyNet        int
	MonthlySpending int
	MonthlyIncome   int
	MonthlyNet      int
	YearlySpending  int
	YearlyIncome    int
	YearlyNet       int
}

func CalculateStats(results []Result) TXStats {
	count := len(results)
	if count <= 1 {
		return TXStats{}
	}

	ci := count - 1

	// // Cumulative expenses at the end of the calculation period.
	var cuex int

	// // Cumulative income at the end of the calculation period.
	var cuin int

	// Daily spending average.
	var ds int

	// Daily income average.
	var di int

	// Monthly spending average.
	var ms int

	// Monthly income average.
	var mi int

	// Year spending average.
	var ys int

	// Year income average.
	var yi int

	cuex = results[ci].CumulativeExpenses
	cuin = results[ci].CumulativeIncome

	ds = CalculateDailyRate(cuex, count)
	di = CalculateDailyRate(cuin, count)
	ms = CalculateMonthlyRate(cuex, count)
	mi = CalculateMonthlyRate(cuin, count)
	ys = CalculateYearlyRate(cuex, count)
	yi = CalculateYearlyRate(cuin, count)

	return TXStats{
		DailySpending:   ds,
		DailyIncome:     di,
		DailyNet:        ds + di,
		MonthlySpending: ms,
		MonthlyIncome:   mi,
		MonthlyNet:      ms + mi,
		YearlySpending:  ys,
		YearlyIncome:    yi,
		YearlyNet:       ys + yi,
	}
}

func (s *TXStats) GetStats() string {
	return fmt.Sprintf(`Here are some statistics about your finances.

Daily spending: %v
Daily income: %v
Daily net: %v
Monthly spending: %v
Monthly income: %v
Monthly net: %v
Yearly spending: %v
Yearly income: %v
Yearly net: %v`,
		FormatAsCurrency(s.DailySpending),
		FormatAsCurrency(s.DailyIncome),
		FormatAsCurrency(s.DailyNet),
		FormatAsCurrency(s.MonthlySpending),
		FormatAsCurrency(s.MonthlyIncome),
		FormatAsCurrency(s.MonthlyNet),
		FormatAsCurrency(s.YearlySpending),
		FormatAsCurrency(s.YearlyIncome),
		FormatAsCurrency(s.YearlyNet),
	)
}

// GetStats spits out some quick calculations about the provided set of results.
// Calculations include, for example, yearly+monthly+daily income/expenses, as
// well as some other things. Users may want to copy this information to the
// clipboard.
func GetStats(results []Result) string {
	s := CalculateStats(results)

	return s.GetStats()
}

func GetResultsCSVString(results *[]Result) string {
	b := new(strings.Builder)
	w := csv.NewWriter(b)

	for _, r := range *results {
		var record []string
		record = append(record, GetNowDateString(r.Date))
		record = append(record, FormatAsCurrency(r.Balance))
		record = append(record, FormatAsCurrency(r.CumulativeIncome))
		record = append(record, FormatAsCurrency(r.CumulativeExpenses))
		record = append(record, FormatAsCurrency(r.DayExpenses))
		record = append(record, FormatAsCurrency(r.DayIncome))
		record = append(record, FormatAsCurrency(r.DayNet))
		record = append(record, FormatAsCurrency(r.DiffFromStart))
		record = append(record, r.DayTransactionNames)
		_ = w.Write(record)
	}

	w.Flush()

	return b.String()
}

// func GetUser() *user.User {
// 	user, err := user.Current()
// 	if err != nil {
// 		log.Printf("failed to get the user's home directory: %v", err.Error())
// 	}

// 	return user
// }

// GetNextSort takes the current sort, which is typically something like
// OrderAsc, OrderDesc, or None, and attempts to do some basic string parsing
// to figure out what the next sort should be. The cycle is None -> Asc -> Des
// Note that if the `next` argument is a different column than the `current`
// argument (after stripping away Asc/Desc), the resulting sort will always be
// the `next` column with Asc ordering.
func GetNextSort(current, next string) string {
	if next == None {
		return None
	}

	if current == None {
		return fmt.Sprintf("%v%v", next, Asc)
	}

	base := strings.TrimSuffix(current, Desc)
	base = strings.TrimSuffix(base, Asc)

	if strings.HasSuffix(current, Desc) {
		if base != next {
			return fmt.Sprintf("%v%v", next, Asc)
		}

		return None
	}

	if strings.HasSuffix(current, Asc) {
		if base != next {
			return fmt.Sprintf("%v%v", next, Asc)
		}

		return fmt.Sprintf("%v%v", base, Desc)
	}

	return fmt.Sprintf("%v%v", next, Asc)
}
