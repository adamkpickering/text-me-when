package reminder

import (
	"testing"
	"time"
)

// Tests conditions that should result in successful CronTrigger creation.
func TestNewCronTriggerNormal(t *testing.T) {
	regular_test_arguments := []string{"2", "*/2", "2,3", "2,3,4,5,6", "1,18,23"}
	for _, argument := range regular_test_arguments {
		_, err := NewCronTrigger("1", argument, "3", "4", "5")
		if err != nil {
			t.Errorf("arg %s: %s", argument, err)
		}
	}
}

// Tests all of the conditions that should result in a failure to create a CronTrigger.
func TestNewCronTriggerAbnormal(t *testing.T) {
	test_arguments := [][]string{
		// lower bounds
		[]string{"-1", "2", "3", "4", "5"},
		[]string{"1", "-1", "3", "4", "5"},
		[]string{"1", "2", "0", "4", "5"},
		[]string{"1", "2", "3", "0", "5"},
		[]string{"1", "2", "3", "4", "-1"},

		// upper bounds
		[]string{"60", "2", "3", "4", "5"},
		[]string{"1", "24", "3", "4", "5"},
		[]string{"1", "2", "32", "4", "5"},
		[]string{"1", "2", "3", "13", "5"},
		[]string{"1", "2", "3", "4", "7"},

		// floating point numbers
		[]string{"1.1", "2", "3", "4", "5"},
		[]string{"1", "1.1", "3", "4", "5"},
		[]string{"1", "2", "1.1", "4", "5"},
		[]string{"1", "2", "3", "1.1", "5"},
		[]string{"1", "2", "3", "4", "1.1"},

		// badly formatted cron strings
		[]string{"*/ 3", "2", "3", "4", "5"},
		[]string{"4,", "2", "3", "4", "5"},
		[]string{"*3", "2", "3", "4", "5"},
		[]string{"* /3", "2", "3", "4", "5"},
		[]string{"4,5,", "2", "3", "4", "5"},
		[]string{",4", "2", "3", "4", "5"},
	}
	for _, arg_list := range test_arguments {
		_, err := NewCronTrigger(arg_list[0], arg_list[1], arg_list[2], arg_list[3], arg_list[4])
		if err == nil {
			t.Errorf("no error when there should have been with args %s", arg_list)
		}
	}
}

type TPCFArguments struct {
	FieldValue string
	LowerBound uint
	UpperBound uint
}

type TPCFTestCase struct {
	Args           TPCFArguments
	ExpectedReturn []uint
}

// Tests cases that should succeed, including return values.
func TestParseCronFieldNormal(t *testing.T) {
	test_cases := []TPCFTestCase{
		// comma-separated case
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "1,2", LowerBound: 0, UpperBound: 3},
			ExpectedReturn: []uint{1, 2},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "1,2,3,4", LowerBound: 0, UpperBound: 4},
			ExpectedReturn: []uint{1, 2, 3, 4},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "3,4,1,2", LowerBound: 0, UpperBound: 4},
			ExpectedReturn: []uint{3, 4, 1, 2},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "1,10,7", LowerBound: 0, UpperBound: 10},
			ExpectedReturn: []uint{1, 10, 7},
		},

		// slash-separated case
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "*/2", LowerBound: 0, UpperBound: 2},
			ExpectedReturn: []uint{0, 2},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "*/4", LowerBound: 0, UpperBound: 10},
			ExpectedReturn: []uint{0, 4, 8},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "*/4", LowerBound: 1, UpperBound: 10},
			ExpectedReturn: []uint{1, 5, 9},
		},

		// asterisk case
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "*", LowerBound: 1, UpperBound: 2},
			ExpectedReturn: []uint{1, 2},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "*", LowerBound: 1, UpperBound: 4},
			ExpectedReturn: []uint{1, 2, 3, 4},
		},
		TPCFTestCase{
			Args:           TPCFArguments{FieldValue: "*", LowerBound: 12, UpperBound: 14},
			ExpectedReturn: []uint{12, 13, 14},
		},
	}
	for _, tc := range test_cases {
		rv, err := parseCronField(tc.Args.FieldValue, tc.Args.LowerBound, tc.Args.UpperBound)
		if err != nil {
			t.Errorf("got unexpected error: %s", err)
		}
		for i, value := range rv {
			if value != tc.ExpectedReturn[i] {
				t.Errorf("got return value %v with args %v (%v expected)", rv, tc.Args, tc.ExpectedReturn)
			}
		}
	}
}

// Tests cases that should fail.
func TestParseCronFieldAbnormal(t *testing.T) {
	test_arguments := []TPCFArguments{
		// comma-separated case
		TPCFArguments{FieldValue: "0,2", LowerBound: 1, UpperBound: 3},
		TPCFArguments{FieldValue: "0,2", LowerBound: 0, UpperBound: 1},
		TPCFArguments{FieldValue: "0,2", LowerBound: 2, UpperBound: 0},

		// slash-separated case
		TPCFArguments{FieldValue: "*/2", LowerBound: 2, UpperBound: 0},

		// asterisk case
		TPCFArguments{FieldValue: "*", LowerBound: 2, UpperBound: 0},
	}
	for _, args := range test_arguments {
		_, err := parseCronField(args.FieldValue, args.LowerBound, args.UpperBound)
		if err == nil {
			t.Errorf("no error where there should have been with args %v", args)
		}
	}
}

func getCronTrigger(t *testing.T, minute, hour, day_of_month, month, day_of_week string) *CronTrigger {
	t.Helper()
	ct, err := NewCronTrigger(minute, hour, day_of_month, month, day_of_week)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
	}
	return ct
}

func TestShouldRun(t *testing.T) {
	// this should always work
	ct := getCronTrigger(t, "*", "*", "*", "*", "*")
	current_time := time.Now()
	if ! ct.ShouldRun(current_time) {
		t.Error("CronTrigger.ShouldRun returned false when it should be true")
	}

	// slash separated 1
	ct = getCronTrigger(t, "*/4", "*", "*", "*", "*")
	test_time := time.Date(2021, time.January, 1, 0, 0, 22, 1234, time.UTC)
	if ! ct.ShouldRun(test_time) {
		t.Error("CronTrigger.ShouldRun returned false when it should be true")
	}

	// slash separated 2
	ct = getCronTrigger(t, "*/4", "*", "*", "*", "*")
	test_time = time.Date(2021, time.January, 1, 0, 8, 22, 1234, time.UTC)
	if ! ct.ShouldRun(test_time) {
		t.Error("CronTrigger.ShouldRun returned false when it should be true")
	}

	// slash separated 3
	ct = getCronTrigger(t, "*/4", "*", "*", "*", "*")
	test_time = time.Date(2021, time.January, 1, 0, 6, 22, 1234, time.UTC)
	t.Log(test_time)
	if ct.ShouldRun(test_time) {
		t.Error("CronTrigger.ShouldRun returned true when it should be false")
	}

	// comma separated 1
	ct = getCronTrigger(t, "*", "0,12,23", "*", "*", "*")
	test_time = time.Date(2021, time.January, 1, 0, 0, 22, 1234, time.UTC)
	t.Log(test_time)
	if ! ct.ShouldRun(test_time) {
		t.Error("CronTrigger.ShouldRun returned false when it should be true")
	}

	// comma separated 2
	ct = getCronTrigger(t, "*", "0,12,23", "*", "*", "*")
	test_time = time.Date(2021, time.January, 1, 12, 0, 22, 1234, time.UTC)
	t.Log(test_time)
	if ! ct.ShouldRun(test_time) {
		t.Error("CronTrigger.ShouldRun returned false when it should be true")
	}

	// comma separated 3
	ct = getCronTrigger(t, "*", "0,12,23", "*", "*", "*")
	test_time = time.Date(2021, time.January, 1, 17, 0, 22, 1234, time.UTC)
	t.Log(test_time)
	if ct.ShouldRun(test_time) {
		t.Error("CronTrigger.ShouldRun returned true when it should be false")
	}
}
