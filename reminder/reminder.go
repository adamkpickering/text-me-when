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
	trigger_type string
	Minute string
	Hour string
	DayOfMonth string
	Month string
	DayOfWeek string
}

func (ct CronTrigger) Type() string {
	return ct.trigger_type
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

func NewCronTrigger(minute, hour, day_of_month, month, day_of_week string) (CronTrigger, error) {
	matched_list_format, err := regexp.MatchString(`[0-9]{1,3}(,[0-9]{1,3})*`, minute)
	if err != nil {
		return CronTrigger{}, errors.New("there was a problem with running regex on minute")
	}
	matched_slash_format, err := regexp.MatchString(`\*/[0-9]{1,2}`, minute)
	if err != nil {
		return CronTrigger{}, errors.New("there was a problem with running regex on minute")
	}
	switch {
	case matched_list_format:
		numbers := strings.Split(minute, ",")
		for _, number := range numbers {
			if ! convertAndCheckBounds(number, 0, 59) {
				return CronTrigger{}, errors.New("field minute: %s is not a valid cron pattern")
			}
		}
	case matched_slash_format:
		split_minute := strings.Split(minute, "/")
		if len(split_minute) != 2 {
			return CronTrigger{}, errors.New("field minute: %s is not a valid cron pattern")
		}
		if ! convertAndCheckBounds(split_minute[1], 0, 59) {
			return CronTrigger{}, errors.New("field minute: %s is not a valid cron pattern")
		}
	case minute == "*":
		break
	default:
		return CronTrigger{}, errors.New("field minute: %s is not a valid cron pattern")
	}

	fields_to_check := map[string]string{
		"minute": minute,
		"hour": hour,
		"day_of_month": day_of_month,
		"month": month,
		"day_of_week": day_of_week,
	}
	pattern := `(\*/[0-9]{1,3})|([0-9]{1,3}(,[0-9]{1,3})*)`
	for field_name, field_value := range fields_to_check {
		matched, _ := regexp.MatchString(pattern, field_value)
		if ! matched {
			msg := fmt.Sprintf("field %s: %s is not a valid cron pattern", field_name, field_value)
			return CronTrigger{}, errors.New(msg)
		}
	}

	return_ct := CronTrigger{
		trigger_type: "cron",
		Minute: minute,
		Hour: hour,
		DayOfMonth: day_of_month,
		Month: month,
		DayOfWeek: day_of_week,
	}
	return return_ct, nil
}

func parseCronField(field, field_name string, lower_bound uint, upper_bound uint) ([]uint, error) {
	matched_comma_format, err := regexp.MatchString(`[0-9]{1,3}(,[0-9]{1,3})*`, field)
	if err != nil {
		msg := fmt.Sprintf("there was a problem running comma regex on %s", field_name)
		return CronTrigger{}, errors.New(msg)
	}
	matched_slash_format, err := regexp.MatchString(`\*/[0-9]{1,2}`, field)
	if err != nil {
		msg := fmt.Sprintf("there was a problem running slash regex on %s", field_name)
		return CronTrigger{}, errors.New(msg)
	}
	switch {
	case matched_comma_format:
		numbers := strings.Split(field, ",")
		for _, number := range numbers {
			if ! convertAndCheckBounds(number, 0, 59) {
				msg := fmt.Sprintf("field %s: %s is not a valid cron pattern", field_name, field)
				return CronTrigger{}, errors.New(msg)
			}
		}
	case matched_slash_format:
		split_field := strings.Split(field, "/")
		if len(split_field) != 2 {
			msg := fmt.Sprintf("field %s: %s is not a valid cron pattern", field_name, field)
			return CronTrigger{}, errors.New(msg)
		}
		if ! convertAndCheckBounds(split_minute[1], 0, 59) {
			msg := fmt.Sprintf("field %s: %s is not a valid cron pattern", field_name, field)
			return CronTrigger{}, errors.New(msg)
		}
	case minute == "*":
		break
	default:
		return CronTrigger{}, errors.New("field minute: %s is not a valid cron pattern")
	}
	// generate list of numbers
}

// Checks that a value can be converted to uint and that it is in
// the range [min, max].
func convertAndCheckBounds(str_value string, min uint, max uint) bool {
	raw_value, err := strconv.ParseUint(str_value, 10, 32)
	value := uint(raw_value)
	if err != nil { return false }
	if value < min { return false }
	if value > max { return false }
	return true
}
