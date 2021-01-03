package reminder

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"encoding/json"
)

// bounds gives the minimum and maximum acceptable values for each cron field.
// These bounds are inclusive. For example, both 0 and 59 are acceptable values
// for the "minute" field.
var bounds = map[string]map[string]uint{
	"minute": map[string]uint{
		"lower": 0,
		"upper": 59,
	},
	"hour": map[string]uint{
		"lower": 0,
		"upper": 23,
	},
	"day_of_month": map[string]uint{
		"lower": 1,
		"upper": 31,
	},
	"month": map[string]uint{
		"lower": 1,
		"upper": 12,
	},
	"day_of_week": map[string]uint{
		"lower": 0,
		"upper": 6,
	},
}

// CronTrigger is a type of Trigger that mimics the behaviour of cron.
// For the uninitiated, each field corresponds to a level of timekeeping:
// minutes, hours, days of the month, months, and days of the week.
// Positive matches for each field are ANDed, with the exception of the
// DayOfMonth and DayOfWeek fields, which are ORed. Each field may use one
// of three formats:
//
// "*": the field matches every value
//
// "*/x": the field matches numbers starting at the low bound of the field
// and counting up by increments of x. So if the low bound is 0 and x = 4,
// the field will match on values 0, 4, 8, 12, 16, and so on.
//
// "x,y,z": the field matches on exactly x, y and z. You may include any number
// of values.
type CronTrigger struct {
	triggerType string
	Minute      string
	Hour        string
	DayOfMonth  string
	Month       string
	DayOfWeek   string
}

// Returns the type of the Trigger.
func (ct *CronTrigger) TriggerType() string {
	return ct.triggerType
}

// Given a time as a time.Time object, tells the caller whether the CronTrigger
// should run at this time.
func (ct *CronTrigger) ShouldRun(current_time time.Time) bool {
	minute := matchCronFields(uint(current_time.Minute()), ct.Minute,
		bounds["minute"]["lower"], bounds["minute"]["upper"])
	hour := matchCronFields(uint(current_time.Hour()), ct.Hour,
		bounds["hour"]["lower"], bounds["hour"]["upper"])
	day_of_month := matchCronFields(uint(current_time.Day()), ct.DayOfMonth,
		bounds["day_of_month"]["lower"], bounds["day_of_month"]["upper"])
	month := matchCronFields(uint(current_time.Month()), ct.Month,
		bounds["month"]["lower"], bounds["month"]["upper"])
	day_of_week := matchCronFields(uint(current_time.Weekday()), ct.DayOfWeek,
		bounds["day_of_week"]["lower"], bounds["day_of_week"]["upper"])

	return minute && hour && month && (day_of_month || day_of_week)
}

// Parses a []byte containing JSON into a CronTrigger.
func (ct *CronTrigger) UnmarshalJSON(data []byte) error {
	if string(data) == "null" { return nil }
	obj := map[string]string{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return fmt.Errorf("problem unmarshalling json: %w", err)
	}
	return ct.mapToCronTrigger(obj)
}

// Parses a map[string]interface{} into a CronTrigger. This is used when unmarshalling
// (from JSON) structs that include a CronTrigger under a field.
func (ct *CronTrigger) ParseTriggerFromInterfaceMap(raw_obj_map map[string]interface{}) error {
	obj_map := map[string]string{}
	for key, i := range raw_obj_map {
		value, ok := i.(string)
		if ! ok {
			return fmt.Errorf("the value of key \"%s\" could not be converted to string", key)
		}
		obj_map[key] = value
	}
	return ct.mapToCronTrigger(obj_map)
}

// Takes a map[string]string and parses its keys and values into their appropriate
// locations in a CronTrigger struct.
func (ct *CronTrigger) mapToCronTrigger(obj map[string]string) error {
	for key, value := range obj {
		switch key {
		case "trigger_type":
			if value != "cron" {
				return fmt.Errorf("trigger type \"value\" is not valid (must be \"cron\")")
			}
			ct.triggerType = value
		case "minute":
			_, err := parseCronField(value, bounds["minute"]["lower"], bounds["minute"]["upper"])
			if err != nil { return generateError(key, value) }
			ct.Minute = value
		case "hour":
			_, err := parseCronField(value, bounds["hour"]["lower"], bounds["hour"]["upper"])
			if err != nil { return generateError(key, value) }
			ct.Hour = value
		case "day_of_month":
			_, err := parseCronField(value, bounds["day_of_month"]["lower"], bounds["day_of_month"]["upper"])
			if err != nil { return generateError(key, value) }
			ct.DayOfMonth = value
		case "month":
			_, err := parseCronField(value, bounds["month"]["lower"], bounds["month"]["upper"])
			if err != nil { return generateError(key, value) }
			ct.Month = value
		case "day_of_week":
			_, err := parseCronField(value, bounds["day_of_week"]["lower"], bounds["day_of_week"]["upper"])
			if err != nil { return generateError(key, value) }
			ct.DayOfWeek = value
		default:
			return fmt.Errorf("the key \"%s\" is not a valid key", key)
		}
	}
	return nil
}

// This is a convenience function to cut down boilerplate in error handling in
// *CronTrigger.mapToCronTrigger.
func generateError(key, value string) error {
	return fmt.Errorf("pattern \"%s\" is invalid for key \"%s\"", value, key)
}

// Checks whether a cron field with a given value matches a field pattern, as stored by
// the CronTrigger object. lower_bound and upper_bound are the bounds (inclusive) of
// the field. Errors in parseCronField are ignored since any problems here should be dealt
// with upon CronTrigger creation.
func matchCronFields(field_value uint, field_pattern string, lower_bound, upper_bound uint) bool {
	numbers, err := parseCronField(field_pattern, lower_bound, upper_bound)
	if err != nil {
		return false
	}
	for _, number := range numbers {
		if number == field_value {
			return true
		}
	}
	return false
}

// Creates a new CronTrigger object from arguments that correspond to cron fields.
func NewCronTrigger(minute, hour, day_of_month, month, day_of_week string) (*CronTrigger, error) {
	fields_to_check := map[string]string{
		"minute":       minute,
		"hour":         hour,
		"day_of_month": day_of_month,
		"month":        month,
		"day_of_week":  day_of_week,
	}
	for field_name, field_pattern := range fields_to_check {
		_, err := parseCronField(field_pattern, bounds[field_name]["lower"], bounds[field_name]["upper"])
		if err != nil {
			return nil, fmt.Errorf("NewCronTrigger error on field %s: %w", field_name, err)
		}
	}
	return_ct := &CronTrigger{
		triggerType: "cron",
		Minute:      minute,
		Hour:        hour,
		DayOfMonth:  day_of_month,
		Month:       month,
		DayOfWeek:   day_of_week,
	}
	return return_ct, nil
}

// Parses a single cron field (can be any field) and returns a slice that contains
// all values that the cron-formatted field stood in for. For example, a value of
// "1,3,7" is turned into []uint{1, 3, 7}, and a value of "*/2" is turned into
// []uint{0, 2} for lower bound 0 and upper bound 2.
func parseCronField(field_pattern string, lower_bound uint, upper_bound uint) ([]uint, error) {
	// check args
	if lower_bound >= upper_bound {
		err := fmt.Errorf("lower_bound (value %d) must be lower than upper_bound (value %d)",
			lower_bound, upper_bound)
		return nil, err
	}

	// check if we're working with comma format ("4,5,23") and act accordingly
	matched_comma_format, err := regexp.MatchString(`^[0-9]{1,3}(,[0-9]{1,3})*$`, field_pattern)
	if err != nil {
		msg := fmt.Sprintf("there was a problem running comma regex on %s", field_pattern)
		return nil, errors.New(msg)
	}
	if matched_comma_format {
		string_numbers := strings.Split(field_pattern, ",")
		numbers := make([]uint, 0, len(string_numbers))
		for _, number := range string_numbers {
			value, err := convertAndCheckBounds(number, lower_bound, upper_bound)
			if err != nil {
				msg := fmt.Sprintf("failed to convert value %s: %s", field_pattern, err)
				return nil, errors.New(msg)
			}
			numbers = append(numbers, value)
		}
		return numbers, nil
	}

	// check if we're working with star-slash format ("*/2") and act accordingly
	matched_slash_format, err := regexp.MatchString(`^\*/[0-9]{1,2}$`, field_pattern)
	if err != nil {
		msg := fmt.Sprintf("there was a problem running slash regex on %s", field_pattern)
		return nil, errors.New(msg)
	}
	if matched_slash_format {
		split_field := strings.Split(field_pattern, "/")
		if len(split_field) != 2 {
			msg := fmt.Sprintf("split: %s is not a valid cron pattern", field_pattern)
			return nil, errors.New(msg)
		}
		value, err := convertAndCheckBounds(split_field[1], lower_bound, upper_bound)
		if err != nil {
			msg := fmt.Sprintf("validation of pattern %s failed: %s", field_pattern, err)
			return nil, errors.New(msg)
		}
		numbers := make([]uint, 0, upper_bound-lower_bound+1)
		for i := lower_bound; i <= upper_bound; i = i + value {
			numbers = append(numbers, i)
		}
		return numbers, nil
	}

	// handle lone * case
	if field_pattern == "*" {
		numbers := make([]uint, 0, upper_bound-lower_bound+1)
		for i := lower_bound; i <= upper_bound; i = i + 1 {
			numbers = append(numbers, i)
		}
		return numbers, nil
	}

	// runs if no other case matched
	return nil, errors.New("field %s: %s is not a valid cron pattern")
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
