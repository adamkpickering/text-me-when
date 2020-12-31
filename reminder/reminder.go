// Contains data models.
package reminder

import (
	"time"
	"strings"
	"regexp"
	"errors"
	"fmt"
	"strconv"
)

type Trigger interface {
	Type() string
	ShouldRun(current_time time.Time) bool
}

type ReminderV1 struct {
	Version string
	Message string
	Triggers []Trigger
}

type CronTrigger struct {
	triggerType string
	Minute string
	Hour string
	DayOfMonth string
	Month string
	DayOfWeek string
}

func (ct CronTrigger) Type() string {
	return ct.triggerType
}

func (ct CronTrigger) ShouldRun(current_time time.Time) bool {
	// match minute
	switch {
	case strings.Contains(ct.Minute, "/"):
		return false
	case strings.Contains(ct.Minute, ","):
		return false
	case ct.Minute == "*":
		return false
	}

	// match hour
	// match day of month
	// match month
	// match day of week
	return true
}

type fieldData struct {
	Value string
	LowerBound uint
	UpperBound uint
}

// Creates a new CronTrigger object from arguments that correspond to cron fields.
func NewCronTrigger(minute, hour, day_of_month, month, day_of_week string) (CronTrigger, error) {
	fields_to_check := map[string]fieldData{
		"minute": fieldData{Value: minute, LowerBound: 0, UpperBound: 59},
		"hour": fieldData{Value: hour, LowerBound: 0, UpperBound: 23},
		"day_of_month": fieldData{Value: day_of_month, LowerBound: 1, UpperBound: 31},
		"month": fieldData{Value: month, LowerBound: 1, UpperBound: 12},
		"day_of_week": fieldData{Value: day_of_week, LowerBound: 0, UpperBound: 6},
	}
	for field_name, field_data := range fields_to_check {
		_, err := parseCronField(field_name, field_data.Value, field_data.LowerBound, field_data.UpperBound)
		if err != nil {
			return CronTrigger{}, fmt.Errorf("NewCronTrigger: %w", err)
		}
	}
	return_ct := CronTrigger{
		triggerType: "cron",
		Minute: minute,
		Hour: hour,
		DayOfMonth: day_of_month,
		Month: month,
		DayOfWeek: day_of_week,
	}
	return return_ct, nil
}

// Parses a single cron field (can be any field) and returns a slice that contains
// all values that the cron-formatted field stood in for.
func parseCronField(field_name, field_value string, lower_bound uint, upper_bound uint) ([]uint, error) {
	// check if we're working with comma format ("4,5,23")
	matched_comma_format, err := regexp.MatchString(`^[0-9]{1,3}(,[0-9]{1,3})*$`, field_value)
	if err != nil {
		msg := fmt.Sprintf("there was a problem running comma regex on %s", field_name)
		return nil, errors.New(msg)
	}

	// check if we're working with star-slash format ("*/2")
	matched_slash_format, err := regexp.MatchString(`^\*/[0-9]{1,2}$`, field_value)
	if err != nil {
		msg := fmt.Sprintf("there was a problem running slash regex on %s", field_name)
		return nil, errors.New(msg)
	}

	// validate and generate []uint of valid numbers depending on what format we got
	switch {
	case matched_comma_format:
		string_numbers := strings.Split(field_value, ",")
		numbers := make([]uint, len(string_numbers))
		for _, number := range string_numbers {
			value, err := convertAndCheckBounds(number, lower_bound, upper_bound)
			if err != nil{
				msg := fmt.Sprintf("failed to convert field %s value %s: %s", field_name, field_value, err)
				return nil, errors.New(msg)
			}
			numbers = append(numbers, value)
		}
		return numbers, nil

	case matched_slash_format:
		split_field := strings.Split(field_value, "/")
		if len(split_field) != 2 {
			msg := fmt.Sprintf("field %s: %s is not a valid cron pattern", field_name, field_value)
			return nil, errors.New(msg)
		}
		value, err := convertAndCheckBounds(split_field[1], lower_bound, upper_bound)
		if err != nil {
			msg := fmt.Sprintf("failed to convert field %s value %s: %s", field_name, field_value, err)
			return nil, errors.New(msg)
		}
		numbers := make([]uint, upper_bound - lower_bound + 1)
		for i := lower_bound; i <= upper_bound; i = i + value {
			numbers = append(numbers, i)
		}
		return numbers, nil

	case field_value == "*":
		numbers := make([]uint, upper_bound - lower_bound + 1)
		for i := lower_bound; i <= upper_bound; i = i + 1 {
			numbers = append(numbers, i)
		}
		return numbers, nil

	default:
		return nil, errors.New("field %s: %s is not a valid cron pattern")
	}
}

// Checks that a value can be converted to uint and that it is in
// the range [min, max].
func convertAndCheckBounds(str_value string, min uint, max uint) (uint, error) {
	raw_value, err := strconv.ParseUint(str_value, 10, 32)
	value := uint(raw_value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to uint", str_value)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("converted value was not in range [%d, %d]", min, max)
	}
	return value, nil
}
