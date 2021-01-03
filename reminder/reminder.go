package reminder

import (
	"time"
	"fmt"
	"encoding/json"
)

// The ShouldRun interface is implemented on types that contain data
// that says when something should happen. A time (presumably the current
// time) is passed to the ShouldRun method, and the type decides whether
// its internal data matches that time.
type ShouldRun interface {
	ShouldRun(current_time time.Time) bool
}

// A Reminder is a single object that represents an even that you want
// to be reminded of.
type Reminder interface {
	UnmarshalJSON([]byte) error
	ShouldRun
}

// A Trigger is included as part of a Reminder. It is the part of the
// Reminder that specifies when the message in the Reminder should be sent.
type Trigger interface {
	TriggerType() string
	ParseTriggerFromInterfaceMap(map[string]interface{}) error
	ShouldRun
}

// This is version 1 of the Reminder.
type ReminderV1 struct {
	Version  string
	Message  string
	Triggers []Trigger
}

// Determines whether r.Message should be sent.
func (r *ReminderV1) ShouldRun(current_time time.Time) bool {
	for _, trigger := range r.Triggers {
		if trigger.ShouldRun(current_time) {
			return true
		}
	}
	return false
}

// Unmarshals a []byte of data into a ReminderV1.
func (r *ReminderV1) UnmarshalJSON(data []byte) error {
	if string(data) == "null" { return nil }
	obj := map[string]interface{}{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println(obj)
		return fmt.Errorf("inital unmarshal failed: %w", err)
	}
	for key, i := range obj {
		switch key {
		case "version":
			value, ok := i.(string)
			if ! ok {
				msg := "failed to parse value of key \"version\" into string"
				return fmt.Errorf(msg)
			}
			r.Version = value

		case "message":
			value, ok := i.(string)
			if ! ok {
				msg := "failed to parse value of key \"message\" into string"
				return fmt.Errorf(msg)
			}
			r.Message = value

		case "triggers":
			interface_list, ok := i.([]interface{})
			if ! ok {
				msg := "failed to parse value of key \"triggers\" into []interface{}"
				return fmt.Errorf(msg)
			}
			r.Triggers = make([]Trigger, 0)
			for _, i := range interface_list {
				obj_map, ok := i.(map[string]interface{})
				if ! ok {
					msg := "failed to parse a trigger into a map[string]interface{}"
					return fmt.Errorf(msg)
				}
				trigger, err := parseTriggerFromInterface(obj_map)
				if err != nil {
					return fmt.Errorf("parseTriggerFromInterface: %w", err)
				}
				r.Triggers = append(r.Triggers, trigger)
			}

		default:
			return fmt.Errorf("ReminderV1.UnmarshalJSON: key %s is invalid", key)
		}
	}
	return nil
}

// Determines the Trigger type, calls the appropriate code to parse the Trigger
// part of the JSON into that type of Trigger, and then casts the resulting object
// into the Trigger type.
func parseTriggerFromInterface(obj_map map[string]interface{}) (Trigger, error) {
	trigger_type, ok := obj_map["trigger_type"].(string)
	if ! ok {
		return nil, fmt.Errorf("could not convert value of \"trigger_type\" key to string")
	}
	switch trigger_type {
	case "cron":
		ct := &CronTrigger{}
		err := ct.ParseTriggerFromInterfaceMap(obj_map)
		if err != nil {
			return nil, fmt.Errorf("CronTrigger.ParseTriggerFromInterfaceMap: %w", err)
		}
		return Trigger(ct), nil
	default:
		return nil, fmt.Errorf("trigger type %s is not a valid trigger type", trigger_type)
	}
}
