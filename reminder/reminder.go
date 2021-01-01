package reminder

import (
	"time"
)

type Trigger interface {
	Type() string
	ShouldRun(current_time time.Time) bool
}

type ReminderV1 struct {
	Version  string
	Message  string
	Triggers []Trigger
}
