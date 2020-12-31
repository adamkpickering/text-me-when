package reminder

import (
	"fmt"
	"testing"
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
	fieldName  string
	fieldValue string
	lowerBound uint
	upperBound uint
}

func (a TPCFArguments) String() string {
	return fmt.Sprintf(`TPCFArguments{fieldName: "%s", fieldValue: "%s", lowerBound: %d, upperBound: %d`,
		a.fieldName, a.fieldValue, a.lowerBound, a.upperBound)
}

// Tests cases that should succeed for comma-separated format.
//func TestParseCronFieldCommaSeparatedNormal(t * testing.T) {
//	test_arguments := []TPCFArguments{
//		TPCFArguments{fieldName: "x", fieldValue: "1,2", lowerBound: 0, upperBound: 3},
//	}
//	for _, args := range test_arguments {
//
//	}
//}

// Tests cases that should fail for comma-separated format.
func TestParseCronFieldCommaSeparatedAbnormal(t *testing.T) {
	test_arguments := []TPCFArguments{
		TPCFArguments{fieldName: "x", fieldValue: "0,2", lowerBound: 1, upperBound: 3},
		TPCFArguments{fieldName: "x", fieldValue: "0,2", lowerBound: 0, upperBound: 1},
		TPCFArguments{fieldName: "x", fieldValue: "0,2", lowerBound: 2, upperBound: 0},
	}
	for _, args := range test_arguments {
		_, err := parseCronField(args.fieldName, args.fieldValue, args.lowerBound, args.upperBound)
		if err == nil {
			t.Errorf("no error where there should have been with args %s", args)
		}
	}
}

// Tests cases that should succeed for slash-separated format.
func TestParseCronFieldSlashSeparatedNormal(t *testing.T) {
}

// Tests cases that should fail for slash-separated format.
func TestParseCronFieldSlashSeparatedAbnormal(t *testing.T) {
}

// Tests cases that should succeed for asterisk ("*") format.
func TestParseCronFieldAsterisk(t *testing.T) {
}
