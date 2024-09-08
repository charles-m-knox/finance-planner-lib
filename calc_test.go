package fplib_test

import (
	"testing"

	fpl "github.com/charles-m-knox/finance-planner-lib"
)

func TestCalculateMonthlyRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount int
		days   int
		want   int
	}{
		{10000, 10, 30438},
		{-10000, 10, -30438},
		{8372302, 15231, 16731},
		{0, 2981, 0},
	}

	for i, test := range tests {
		got := fpl.CalculateMonthlyRate(test.amount, test.days)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}

func TestCalculateYearlyRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount int
		days   int
		want   int
	}{
		{10000, 553, 6605},
		{0, 200, 0},
		{-10000, 553, -6605},
	}

	for i, test := range tests {
		got := fpl.CalculateYearlyRate(test.amount, test.days)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}

func TestCalculateDailyRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount int
		days   int
		want   int
	}{
		{10000, 9231, 1},
		{0, 200, 0},
		{-10000, 9231, -1},
		{-5438210, 302, -18007},
	}

	for i, test := range tests {
		got := fpl.CalculateDailyRate(test.amount, test.days)
		if got != test.want {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.FailNow()
		}
	}
}
