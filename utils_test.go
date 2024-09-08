package fplib_test

import (
	"testing"
	"time"

	fpl "github.com/charles-m-knox/finance-planner-lib"
)

func TestFormatAsCurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input int
		want  string
	}{
		{-1, "$-0.01"},
		{0, "$0.00"},
		{-99, "$-0.99"},
		{-100, "$-1.00"},
		{99, "$0.99"},
		{100, "$1.00"},
		{10000, "$100.00"},
		{-10000, "$-100.00"},
		{-12345, "$-123.45"},
	}

	for i, test := range tests {
		got := fpl.FormatAsCurrency(test.input)
		if test.want != got {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.Fail()
		}
	}
}

func TestGetCSVString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input []string
		want  string
	}{
		{[]string{}, ""},
		{[]string{"1", "2", "3"}, "(3) 1; 2; 3; "},
	}

	for i, test := range tests {
		got := fpl.GetCSVString(test.input)
		if test.want != got {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.Fail()
		}
	}
}

func TestParseDollarAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		positive bool
		want     int64
	}{
		{"$100.00", false, -10000},
		// positive-valued output test cases with assumePositive=false
		{"+$100.00", false, 10000},
		{"$+100.00", false, 10000},
		// positive-valued output test case with assumePositive=true
		{"$100.00", true, 10000},
		// negative-valued output test cases with assumePositive=true
		{"-$100.00", true, -10000},
		{"$-100.00", true, -10000},
		// cases where user entered a period
		{"$100.2", false, -10020},
		{"$100.01", false, -10001},
		{"$100.99", false, -10099},
		// too many numbers get zeroed out - assume it was invalid input
		{"$100.243", false, -10000},
		// test cases regardless of a period's presence
		{"-100", false, -10000},
		{"23", false, -2300},
		{"$23", false, -2300},
		{"$9221", true, 922100},
		// test cases with messy input
		{"asdf3n3201", true, 3320100},
		{"-asdf3n3201", true, -3320100},
		{"-asdf3n3201", true, -3320100},
		{"$-asdf3n3201", true, -3320100},
		{"483#@#**@*^^^.92", false, -48392},
		{"483#@#**@*^^^.92asdflij3m3kci9", false, -48300},
		{"$483#@#**@*^^^.92asdflij3m3kci9", false, -48300},
		{"$-483#@#**@*^^^.92asdflij3m3kci9", false, -48300},
		{"$-483#@#**@*^^^.92asdflij3m3kci9", true, -48300},
		{"$+483#@#**@*^^^.92asdflij3m3kci9", true, 48300},
		{"$+483#@#**@*^^^.92asdflij3m3kci9", false, 48300},
	}

	for i, test := range tests {
		got := fpl.ParseDollarAmount(test.input, test.positive)
		if test.want != got {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.Fail()
		}
	}
}

func TestParseYearMonthDateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want1 int
		want2 int
		want3 int
	}{
		{"2020-01-01", 2020, 1, 1},
		{"2020-0101", 0, 0, 0},
		{"20200101", 0, 0, 0},
		{"---", 0, 0, 0},
		{"-1--1--1", 0, 0, 0},
	}

	for i, test := range tests {
		got1, got2, got3 := fpl.ParseYearMonthDateString(test.input)
		if test.want1 != got1 || test.want2 != got2 || test.want3 != got3 {
			t.Logf("test %v failed: got %v %v %v but wanted %v %v %v", i, got1, got2, got3, test.want1, test.want2, test.want3)
			t.Fail()
		}
	}
}

func TestGetDateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		y, m, d any
		want    string
	}{
		{2024, 12, 1, "2024-12-01"},
		{24, 12, 1, "0024-12-01"},
		{1, 1, 1, "0001-01-01"},
	}

	for i, test := range tests {
		got := fpl.GetDateString(test.y, test.m, test.d)
		if test.want != got {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.Fail()
		}
	}
}

func TestGetDefaultEndDateString(t *testing.T) {
	t.Parallel()

	n := time.Date(2024, 12, 1, 10, 23, 1, 0, time.UTC)

	tests := []struct {
		t    time.Time
		want string
	}{
		{n, "2025-12-01"},
	}

	for i, test := range tests {
		got := fpl.GetDefaultEndDateString(test.t)
		if test.want != got {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.Fail()
		}
	}
}

func TestGetNowDateString(t *testing.T) {
	t.Parallel()

	n := time.Date(2024, 12, 1, 10, 23, 1, 0, time.UTC)

	tests := []struct {
		t    time.Time
		want string
	}{
		{n, "2024-12-01"},
	}

	for i, test := range tests {
		got := fpl.GetNowDateString(test.t)
		if test.want != got {
			t.Logf("test %v failed: got %v but wanted %v", i, got, test.want)
			t.Fail()
		}
	}
}
