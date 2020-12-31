package main

import (
	"time"
	"fmt"
	"flag"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"os"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sns"
)

type Reminder struct {
	FirstName string
	LastName string
	Birthday time.Time
}

func (r *Reminder) UnmarshalJSON(data []byte) error {
	if string(data) == "null" { return nil }
	obj := map[string]string{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return fmt.Errorf("Reminder.UnmarshalJSON: %w", err)
	}
	for key, value := range obj {
		switch key {
		case "first_name":
			r.FirstName = value
		case "last_name":
			r.LastName = value
		case "birthday":
			birthday, err := time.Parse("2006-01-02", value)
			if err != nil {
				return fmt.Errorf("Reminder.UnmarshalJSON: %w", err)
			}
			r.Birthday = birthday
		default:
			msg := fmt.Sprintf("Reminder.UnmarshalJSON: key %s is invalid", key)
			return errors.New(msg)
		}
	}
	return nil
}

func send_message(sns_client *sns.SNS, message string, number string) error {
	pi := &sns.PublishInput{
		Message: &message,
		PhoneNumber: &number,
	}
	if err := pi.Validate(); err != nil {
		return fmt.Errorf("send_message: %w", err)
	}
	fmt.Printf("sending msg \"%s\" to \"%s\"\n", message, number)
	_, err := sns_client.Publish(pi)
	if err != nil {
		return fmt.Errorf("send_message: %w", err)
	}
	return nil
}

func fire_reminders(call_time time.Time, sns_client *sns.SNS, reminder_list []Reminder) {
	test_message := "this is a test message"
	test_number := "+12345678901"

	// filter reminders down to day
	_, call_month, call_day := call_time.Date()
	call_hour := call_time.Hour()
	call_minute := call_time.Minute()
	for _, reminder := range reminder_list {
		_, rem_month, rem_day := reminder.Birthday.Date()
		rem_hour := 9
		rem_minute := 0
		if call_month == rem_month && call_day == rem_day && call_hour == rem_hour && call_minute == rem_minute {
			err := send_message(sns_client, test_message, test_number)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fire_reminders: %w", err)
			}
		}
	}
}

func main() {
	// parse CLI flags
	config_path := flag.String("c", "/etc/reminder-boi.json", "The path to the config file")
	flag.Parse()

	// construct sns client
	home_dir, ok := os.LookupEnv("HOME")
	if ! ok {
		fmt.Fprintf(os.Stderr, "main: HOME env variable is not present")
		os.Exit(1)
	}
	creds_path := filepath.Join(home_dir, ".aws/credentials")
	creds := credentials.NewSharedCredentials(creds_path, "default")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion("ca-central-1")
	session := session.Must(session.NewSession(cfg))
	sns_client := sns.New(session)

	// parse config file
	raw_file, err := ioutil.ReadFile(*config_path)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}
	reminder_list := make([]Reminder, 20)
	err = json.Unmarshal(raw_file, &reminder_list)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}

	// main loop
	var wait_time time.Duration = 60
	ticker := time.NewTicker(wait_time*time.Second)
	for {
		received_time := <-ticker.C
		fire_reminders(received_time, sns_client, reminder_list)
	}
}
