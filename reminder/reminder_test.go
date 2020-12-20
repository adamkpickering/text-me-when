package reminder

import (
	"testing"
)

func TestNewCronTrigger(t *testing.T) {
	// test what should be successful creation
	regular_test_arguments := []string{"2", "*/2", "2,3", "2,3,4,5,6", "24,59"}
	for _, argument := range regular_test_arguments {
		_, err := NewCronTrigger("1", argument, "3", "4", "5")
		if err != nil {
			t.Errorf("regular creation with arg %s failed", argument)
		}
	}
	//
	// look at output
	//

	// test what should be unsuccessful creation
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
		[]string{"1", "2", "3", "4", "5"},
	}
	for _, arg_list := range test_arguments {
		_, err := NewCronTrigger(arg_list[0], arg_list[1], arg_list[2], arg_list[3], arg_list[4])
		if err == nil {
			t.Errorf("no error when there should have been with args %s", arg_list)
		}
	}
}
