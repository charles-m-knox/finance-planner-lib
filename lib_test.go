package fplib_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	fpl "git.cmcode.dev/cmcode/finance-planner-lib"
	"git.cmcode.dev/cmcode/uuid"
	"github.com/teambition/rrule-go"
)

//nolint:cyclop,maintidx,lll
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
		},
		{
			// This test case is the same as the first test case, while also
			// specifying that this transaction can occur on every possible day
			// of the week.
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
					Weekdays: map[int]bool{
						rrule.MO.Day(): true,
						rrule.TU.Day(): true,
						rrule.WE.Day(): true,
						rrule.TH.Day(): true,
						rrule.FR.Day(): true,
						rrule.SA.Day(): true,
						rrule.SU.Day(): true,
						-1:             true,  // test a nonsense input value
						-2:             false, // test a nonsense input value
					},
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
		},
		{
			// This test case has a daily recurrence pattern, as well
			// as a nonsense recurrence pattern that should default to daily.
			[]fpl.TX{
				{
					Amount:      tx1Amount,
					Name:        tx1,
					Active:      true,
					Frequency:   fpl.WEEKLY,
					Interval:    1,
					StartsDay:   1,
					StartsMonth: 1,
					StartsYear:  2024,
					ID:          uuid.New(),
					Weekdays: map[int]bool{
						rrule.MO.Day(): true,
						rrule.TU.Day(): true,
						rrule.WE.Day(): true,
						rrule.TH.Day(): true,
						rrule.FR.Day(): true,
						rrule.SA.Day(): true,
						rrule.SU.Day(): true,
					},
				},
			},
			start,
			end,
			startBalance,
			[]fpl.Result{{
				Balance:            startBalance - 7640000,
				DiffFromStart:      -7640000,
				CumulativeExpenses: -7640000,
				CumulativeIncome:   0,
			}},
			false,
			days,
		},
		{
			// This test case has a yearly recurrence pattern that occurs
			// only twice over a span of two years.
			[]fpl.TX{
				{
					Amount:      tx1Amount,
					Name:        tx1,
					Active:      true,
					Frequency:   rrule.YEARLY.String(),
					Interval:    1,
					StartsDay:   1,
					StartsMonth: 1,
					StartsYear:  2024,
					EndsDay:     31,
					EndsMonth:   12,
					EndsYear:    2025,
					ID:          uuid.New(),
				},
			},
			start,
			end,
			startBalance,
			[]fpl.Result{{
				Balance:            startBalance + 2*tx1Amount,
				DiffFromStart:      2 * tx1Amount,
				CumulativeExpenses: 2 * tx1Amount,
				CumulativeIncome:   0,
			}},
			false,
			days,
		},
		{
			// This test case uses an rrule string that matches the previous
			// test case.
			[]fpl.TX{
				{
					Amount: tx1Amount,
					Name:   tx1,
					Active: true,
					RRule:  "DTSTART:20240101T000000Z\nRRULE:FREQ=YEARLY;INTERVAL=1;UNTIL=20251231T000000Z",
					ID:     uuid.New(),
				},
			},
			start,
			end,
			startBalance,
			[]fpl.Result{{
				Balance:            startBalance + 2*tx1Amount,
				DiffFromStart:      2 * tx1Amount,
				CumulativeExpenses: 2 * tx1Amount,
				CumulativeIncome:   0,
			}},
			false,
			days,
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
	}
}

//nolint:cyclop
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

func TestGetStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		results []fpl.Result
		want    string
	}{
		{
			[]fpl.Result{
				{CumulativeIncome: 10000, CumulativeExpenses: -10000},
				{CumulativeIncome: 20000, CumulativeExpenses: -20000},
			}, `Here are some statistics about your finances.

Daily spending: $-100.00
Daily income: $100.00
Daily net: $0.00
Monthly spending: $-3043.75
Monthly income: $3043.75
Monthly net: $0.00
Yearly spending: $-36525.00
Yearly income: $36525.00
Yearly net: $0.00`,
		},
		{
			[]fpl.Result{
				{CumulativeIncome: 10000, CumulativeExpenses: -10000},
			}, `Here are some statistics about your finances.

Daily spending: $0.00
Daily income: $0.00
Daily net: $0.00
Monthly spending: $0.00
Monthly income: $0.00
Monthly net: $0.00
Yearly spending: $0.00
Yearly income: $0.00
Yearly net: $0.00`,
		},
		{
			[]fpl.Result{}, `Here are some statistics about your finances.

Daily spending: $0.00
Daily income: $0.00
Daily net: $0.00
Monthly spending: $0.00
Monthly income: $0.00
Monthly net: $0.00
Yearly spending: $0.00
Yearly income: $0.00
Yearly net: $0.00`,
		},
	}

	for i, test := range tests {
		got := fpl.GetStats(test.results)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}

func TestGetResultsCSVString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		results *[]fpl.Result
		want    string
	}{
		{
			&[]fpl.Result{
				{CumulativeIncome: 10000, CumulativeExpenses: -10000},
				{CumulativeIncome: 20000, CumulativeExpenses: -20000},
			}, `0001-01-01,$0.00,$100.00,$-100.00,$0.00,$0.00,$0.00,$0.00,
0001-01-01,$0.00,$200.00,$-200.00,$0.00,$0.00,$0.00,$0.00,
`,
		},
	}

	for i, test := range tests {
		got := fpl.GetResultsCSVString(test.results)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}

func TestGetDateFromStrSafe(t *testing.T) {
	t.Parallel()

	now := time.Now()
	t1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tests := []struct {
		s    string
		t    time.Time
		want time.Time
	}{
		{"0-0-0", now, t1},
		{"--", now, t1},
		{"", now, t1},
		{"2024-02-01", now, time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)},
	}

	for i, test := range tests {
		got := fpl.GetDateFromStrSafe(test.s, test.t)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}

// Tests various minimal utility functions relating to transactions.
//
//nolint:cyclop
func TestTXUtilityFunctions(t *testing.T) {
	t.Parallel()

	tx1 := fpl.TX{
		StartsYear:  2024,
		StartsMonth: int(time.January),
		StartsDay:   1,
		EndsYear:    2025,
		EndsMonth:   int(time.February),
		EndsDay:     2,
		ID:          "foo1",
	}

	tx2 := fpl.TX{
		StartsYear:  2024,
		StartsMonth: int(time.January),
		StartsDay:   1,
		EndsYear:    2025,
		EndsMonth:   int(time.February),
		EndsDay:     2,
		ID:          "foo2",
	}

	tx3 := fpl.TX{
		StartsYear:  2024,
		StartsMonth: int(time.January),
		StartsDay:   1,
		EndsYear:    2025,
		EndsMonth:   int(time.February),
		EndsDay:     2,
		ID:          "foo3",
	}

	{
		got := tx1.GetStartDateString()
		want := fpl.GetDateString(tx1.StartsYear, tx1.StartsMonth, tx1.StartsDay)

		if got != want {
			t.Logf("tx1.GetStartDateString test failed: got %v, want %v", got, want)
			t.Fail()
		}
	}

	{
		got := tx1.GetEndsDateString()
		want := fpl.GetDateString(tx1.EndsYear, tx1.EndsMonth, tx1.EndsDay)

		if got != want {
			t.Logf("tx1.GetEndsDateString test failed: got %v, want %v", got, want)
			t.Fail()
		}
	}

	{
		got := fpl.RemoveTXAtIndex([]fpl.TX{tx1, tx2, tx3}, 0)
		want := []fpl.TX{tx2, tx3}

		if len(got) != len(want) {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", len(got), len(want))
			t.FailNow()
		}

		if got[0].ID != want[0].ID {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", got[0].ID, want[0].ID)
			t.FailNow()
		}
	}

	{
		got := []fpl.TX{tx1, tx2, tx3}
		want := []fpl.TX{tx2, tx3}

		fpl.RemoveTXByID(&got, "foo1")

		if len(got) != len(want) {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", len(got), len(want))
			t.FailNow()
		}

		if got[0].ID != want[0].ID {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", got[0].ID, want[0].ID)
			t.FailNow()
		}
	}

	{
		txs := []fpl.TX{tx1, tx2, tx3}
		got := tx1
		want := tx1
		wantIndex := 2

		index, err := fpl.GetTXByID(&txs, "foo3")
		if err != nil || index == -1 {
			t.Logf("RemoveTXAtIndex test failed due to error or no index of tx by id")
			t.FailNow()
		}

		if got.ID != want.ID {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", got.ID, want.ID)
			t.FailNow()
		}

		if index != wantIndex {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", index, wantIndex)
			t.FailNow()
		}
	}

	{
		txs := []fpl.TX{tx1, tx2, tx3}

		index, err := fpl.GetTXByID(&txs, "bar") // not present
		if err == nil || index != -1 {
			t.Logf("RemoveTXAtIndex test failed: got %v, want %v", index, -1)
			t.FailNow()
		}
	}
}
