package fplib_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	fpl "git.cmcode.dev/cmcode/finance-planner-lib"
	"git.cmcode.dev/cmcode/uuid"
)

//nolint:lll
func TestGetResults(t *testing.T) {
	t.Parallel()

	statusHook := func(_ string) {}

	const tx1 = "Foo"

	const startBalance = 10000

	start := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.February, 2, 0, 0, 0, 0, time.UTC)

	// TODO: ensure there aren't any off-by-one errors
	days := int(math.Ceil(end.Sub(start).Hours()/24 + 1))
	t.Logf("days=%v", days)

	// The amount for the primary test case every month / interval.
	const tx1Amount int = -10000
	// The expected cost for the primary test case.
	const expectedCostCase1 int = tx1Amount * (12*2 /* 2 years */ + 2 /* extra months in 2026 */)
	// This test case changes the expected cost timeframe to start when
	// the calculation start date begins (2020), as opposed to the usual start date
	// for our primary test case's transaction (2024).
	const expectedCostCase3 int = tx1Amount * ((12)*(2026-2020) + 2) // 2 months of transactions in 2026

	tests := []struct {
		tx         []fpl.TX
		start, end time.Time
		balance    int
		want       []fpl.Result
		// Whether to expect an error or not.
		err bool
		// Length of results (number of days) to expect.
		days int
		// If there are no expected or accidental errors, the stats function
		// should be called and computed.
		stats string
	}{
		{
			// The first test case consists of a monthly expense of $100 and no
			// income.
			[]fpl.TX{
				{
					Amount:      tx1Amount,
					Name:        tx1,
					Active:      true,
					Frequency:   fpl.MONTHLY,
					Interval:    1,
					StartsDay:   1,
					StartsMonth: 1,
					StartsYear:  2024,
					ID:          uuid.New(),
				},
			},
			start,
			end,
			startBalance,
			[]fpl.Result{{
				Balance:            startBalance + expectedCostCase1,
				DiffFromStart:      expectedCostCase1,
				CumulativeExpenses: expectedCostCase1,
				CumulativeIncome:   0,
			}},
			false,
			days,
			`Here are some statistics about your finances.

Daily spending: $-1.17
Daily income: $0.00
Daily net: $-1.17,
Monthly spending: $-14.02
Monthly income: $0.00
Monthly net: $-14.02
Yearly spending: $-426.52
Yearly income: $0.00
Yearly net: $-426.52`,
		},
		{
			// The second test case consists of an equal income and expense
			// every month of $100. It also has an inactive transaction.
			[]fpl.TX{
				{
					Amount:      tx1Amount,
					Name:        tx1,
					Active:      true,
					Frequency:   fpl.MONTHLY,
					Interval:    1,
					StartsDay:   1,
					StartsMonth: 1,
					StartsYear:  2024,
					ID:          uuid.New(),
				},
				{
					Amount:      -1 * tx1Amount,
					Name:        tx1,
					Active:      true,
					Frequency:   fpl.MONTHLY,
					Interval:    1,
					StartsDay:   1,
					StartsMonth: 1,
					StartsYear:  2024,
					ID:          uuid.New(),
				},
				{
					Amount:      -1 * tx1Amount,
					Name:        tx1,
					Active:      false,
					Frequency:   fpl.MONTHLY,
					Interval:    1,
					StartsDay:   1,
					StartsMonth: 1,
					StartsYear:  2024,
					ID:          uuid.New(),
				},
			},
			start,
			end,
			startBalance,
			[]fpl.Result{{
				Balance:            startBalance,
				DiffFromStart:      0,
				CumulativeExpenses: expectedCostCase1,
				CumulativeIncome:   expectedCostCase1 * -1,
			}},
			false,
			days,
			`Here are some statistics about your finances.

Daily spending: $-1.17
Daily income: $1.17
Daily net: $0.00,
Monthly spending: $-14.02
Monthly income: $14.02
Monthly net: $0.00
Yearly spending: $-426.52
Yearly income: $426.52
Yearly net: $0.00`,
		},
		{
			// The third test case handles a few edge cases.
			[]fpl.TX{
				{
					Amount:    tx1Amount,
					Name:      tx1,
					Active:    true,
					Frequency: fpl.MONTHLY,
					Interval:  1,
					// Edge case: If the start day/mo/yr aren't specified,
					// it should assume that the transaction starts at the
					// same time as the start date for the entire calculation
					// StartsDay:   1,
					// StartsMonth: 1,
					// StartsYear:  2024,
					ID: uuid.New(),
				},
			},
			start,
			end,
			startBalance,
			[]fpl.Result{{
				Balance:            startBalance + expectedCostCase3,
				DiffFromStart:      expectedCostCase3,
				CumulativeExpenses: expectedCostCase3,
				CumulativeIncome:   0,
			}},
			false,
			days,
			`Here are some statistics about your finances.

Daily spending: $-3.33
Daily income: $0.00
Daily net: $-3.33,
Monthly spending: $-39.91
Monthly income: $0.00
Monthly net: $-39.91
Yearly spending: $-1213.93
Yearly income: $0.00
Yearly net: $-1213.93`,
		},
		{
			// The fourth test case is simple and triggers an immediate error.
			[]fpl.TX{},
			time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			0,
			[]fpl.Result{},
			true,
			0,
			"",
		},
	}

	for i, test := range tests {
		got, err := fpl.GetResults(test.tx, test.start, test.end, test.balance, statusHook)
		if err != nil && !test.err {
			t.Logf("test %v threw error when it wasn't supposed to: %v", i, err.Error())
			t.FailNow()
		} else if err != nil && test.err {
			// good test result that expected an error
			continue
		}

		// uncomment if needed for debugging
		// for j, result := range got {
		// 	t.Logf("%v: %v", j, result)
		// }

		if len(got) != test.days {
			t.Logf("test %v wrong day count: got %v, want %v (%v results)", i, len(got), days, len(got))
			t.FailNow()
		}

		// for now, only check the final result
		wl := len(test.want)
		gl := len(got)
		wantBalance := test.want[wl-1].Balance
		gotBalance := got[gl-1].Balance

		if gotBalance != wantBalance {
			t.Logf("test %v wrong balance: got %v, want %v (%v results)", i, gotBalance, wantBalance, len(got))
			t.Fail()
		}

		wantDiffFromStart := test.want[wl-1].DiffFromStart
		gotDiffFromStart := got[gl-1].DiffFromStart

		if gotDiffFromStart != wantDiffFromStart {
			t.Logf("test %v wrong DiffFromStart: got %v, want %v (%v results)", i, gotDiffFromStart, wantDiffFromStart, len(got))
			t.Fail()
		}

		wantCumulativeExpenses := test.want[wl-1].CumulativeExpenses
		gotCumulativeExpenses := got[gl-1].CumulativeExpenses

		if gotCumulativeExpenses != wantCumulativeExpenses {
			t.Logf("test %v wrong CumulativeExpenses: got %v, want %v (%v results)", i, gotCumulativeExpenses, wantCumulativeExpenses, len(got))
			t.Fail()
		}

		wantCumulativeIncome := test.want[wl-1].CumulativeIncome
		gotCumulativeIncome := got[gl-1].CumulativeIncome

		if gotCumulativeIncome != wantCumulativeIncome {
			t.Logf("test %v wrong CumulativeIncome: got %v, want %v (%v results)", i, gotCumulativeIncome, wantCumulativeExpenses, len(got))
			t.Fail()
		}

		// compute stats
		gotstats := fpl.GetStats(got)
		// t.Logf("gotstats: %v", gotstats)
		if gotstats != test.stats {
			t.Logf("test %v wrong stats: got %v, want %v", i, gotstats, test.stats)
			t.Fail()
		}
	}
}

func TestGetNewTX(t *testing.T) {
	t.Parallel()

	n := time.Date(2024, 12, 1, 10, 23, 1, 0, time.UTC)

	oneMonth := n.Add(time.Hour * fpl.HoursInDay * fpl.DaysInMonth)

	want := fpl.TX{
		Amount:      fpl.DefaultTransactionBalance,
		Active:      true,
		Name:        fpl.New,
		Frequency:   fpl.MONTHLY,
		Interval:    1,
		StartsDay:   n.Day(),
		StartsMonth: int(n.Month()),
		StartsYear:  n.Year(),
		EndsDay:     oneMonth.Day(),
		EndsMonth:   int(oneMonth.Month()),
		EndsYear:    oneMonth.Year(),
		ID:          uuid.New(),
		CreatedAt:   n,
		UpdatedAt:   n,
		Note:        "",
		RRule:       "",
		Weekdays:    fpl.GetWeekdaysMap(),
		Selected:    false,
	}

	got := fpl.GetNewTX(n)

	if got.Amount != want.Amount {
		t.Logf("Amount mismatch: got %v, want %v", got.Amount, want.Amount)
		t.FailNow()
	}

	if got.Active != want.Active {
		t.Logf("Active mismatch: got %v, want %v", got.Active, want.Active)
		t.FailNow()
	}

	if got.Name != want.Name {
		t.Logf("Name mismatch: got %v, want %v", got.Name, want.Name)
		t.FailNow()
	}

	if got.Frequency != want.Frequency {
		t.Logf("Frequency mismatch: got %v, want %v", got.Frequency, want.Frequency)
		t.FailNow()
	}

	if got.Interval != want.Interval {
		t.Logf("Interval mismatch: got %v, want %v", got.Interval, want.Interval)
		t.FailNow()
	}

	if got.StartsDay != want.StartsDay {
		t.Logf("StartsDay mismatch: got %v, want %v", got.StartsDay, want.StartsDay)
		t.FailNow()
	}

	if got.StartsMonth != want.StartsMonth {
		t.Logf("StartsMonth mismatch: got %v, want %v", got.StartsMonth, want.StartsMonth)
		t.FailNow()
	}

	if got.StartsYear != want.StartsYear {
		t.Logf("StartsYear mismatch: got %v, want %v", got.StartsYear, want.StartsYear)
		t.FailNow()
	}

	if got.EndsDay != want.EndsDay {
		t.Logf("EndsDay mismatch: got %v, want %v", got.EndsDay, want.EndsDay)
		t.FailNow()
	}

	if got.EndsMonth != want.EndsMonth {
		t.Logf("EndsMonth mismatch: got %v, want %v", got.EndsMonth, want.EndsMonth)
		t.FailNow()
	}

	if got.EndsYear != want.EndsYear {
		t.Logf("EndsYear mismatch: got %v, want %v", got.EndsYear, want.EndsYear)
		t.FailNow()
	}

	// uuid's are not worth testing
	if len(got.ID) != len(want.ID) {
		t.Logf("ID mismatch: got %v, want %v", got.ID, want.ID)
		t.FailNow()
	}

	if got.CreatedAt != want.CreatedAt {
		t.Logf("CreatedAt mismatch: got %v, want %v", got.CreatedAt, want.CreatedAt)
		t.FailNow()
	}

	if got.UpdatedAt != want.UpdatedAt {
		t.Logf("UpdatedAt mismatch: got %v, want %v", got.UpdatedAt, want.UpdatedAt)
		t.FailNow()
	}

	if got.Note != want.Note {
		t.Logf("Note mismatch: got %v, want %v", got.Note, want.Note)
		t.FailNow()
	}

	if got.RRule != want.RRule {
		t.Logf("RRule mismatch: got %v, want %v", got.RRule, want.RRule)
		t.FailNow()
	}

	if len(got.Weekdays) != len(want.Weekdays) {
		t.Logf("weekdays len mismatch: got %v, want %v", len(got.Weekdays), len(want.Weekdays))
		t.FailNow()
	}

	for k, v := range want.Weekdays {
		if want.Weekdays[k] != v {
			t.Logf("weekdays mismatch: want k=%v, v=%v but got %v", k, v, want.Weekdays[k])
			t.FailNow()
		}
	}

	if got.Selected != want.Selected {
		t.Logf("Selected mismatch: got %v, want %v", got.Selected, want.Selected)
		t.FailNow()
	}
}

func TestGetWeekdaysCheckedMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tx                 fpl.TX
		checked, unchecked string
		want               map[int]string
	}{
		{fpl.TX{Weekdays: map[int]bool{
			0: true,
			1: true,
			2: false,
			3: false,
			4: true,
			5: true,
			6: true,
		}}, "X", " ", map[int]string{
			0: "X",
			1: "X",
			2: " ",
			3: " ",
			4: "X",
			5: "X",
			6: "X",
		}},
	}

	for i, test := range tests {
		got := test.tx.GetWeekdaysCheckedMap(test.checked, test.unchecked)
		if len(got) != len(test.want) {
			t.Logf("test %v failed: got %v but wanted %v", i, len(got), len(test.want))
			t.FailNow()
		}

		for k, v := range test.want {
			if got[k] != v {
				t.Logf("test %v failed: got %v but wanted %v", i, got[k], k)
				t.FailNow()
			}
		}
	}
}

func TestGetNextSort(t *testing.T) {
	t.Parallel()

	const column1 = "Foo"

	const column2 = "Bar"

	tests := []struct {
		current, next string
		want          string
	}{
		{fpl.None, column1, fmt.Sprintf("%v%v", column1, fpl.Asc)},
		{column1, fpl.None, fpl.None},
		{fpl.None, fpl.None, fpl.None},
		{fmt.Sprintf("%v%v", column1, fpl.Desc), column1, fpl.None},
		{fmt.Sprintf("%v%v", column1, fpl.Desc), column2, fmt.Sprintf("%v%v", column2, fpl.Asc)},
		{fmt.Sprintf("%v%v", column1, fpl.Asc), column1, fmt.Sprintf("%v%v", column1, fpl.Desc)},
		{fmt.Sprintf("%v%v", column1, fpl.Asc), column2, fmt.Sprintf("%v%v", column2, fpl.Asc)},
		{column1, column1, fmt.Sprintf("%v%v", column1, fpl.Asc)},
		{column1, column2, fmt.Sprintf("%v%v", column2, fpl.Asc)},
	}

	for i, test := range tests {
		got := fpl.GetNextSort(test.current, test.next)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}
