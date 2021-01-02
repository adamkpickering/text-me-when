package reminder

import (
	"time"
	"fmt"
	"encoding/json"
)

type ShouldRun interface {
	ShouldRun(current_time time.Time) bool
}

type Reminder interface {
	UnmarshalJSON([]byte) error
	ShouldRun
}

type Trigger interface {
	TriggerType() string
	ParseTriggerFromInterfaceMap(map[string]interface{}) error
	ShouldRun
}

type ReminderV1 struct {
	Version  string
	Message  string
	Triggers []Trigger
}

func (r *ReminderV1) ShouldRun(current_time time.Time) bool {
	for _, trigger := range r.Triggers {
		if trigger.ShouldRun(current_time) {
			return true
		}
	}
	return false
}

func (r *ReminderV1) UnmarshalJSON(data []byte) error {
	if string(data) == "null" { return nil }
	obj := map[string]interface{}{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println(obj)
		return fmt.Errorf("inital unmarshal failed: %w", err)
	}
	if &obj == nil {
		return fmt.Errorf("got null literal")
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
