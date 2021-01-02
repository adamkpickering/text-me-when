package main

import (
	"time"
	"fmt"
	"flag"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/adamkpickering/reminder-boi/reminder"
)

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

func fire_reminders(call_time time.Time, sns_client *sns.SNS, reminder_list []reminder.ReminderV1) {
	test_message := "this is a test message"
	test_number := "+12345678901"
	for _, reminder := range reminder_list {
		if reminder.ShouldRun(call_time) {
			err := send_message(sns_client, test_message, test_number)
			fmt.Printf("sending message \"%s\" to %s...", test_message, test_number)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fire_reminders: %w", err)
			}
			fmt.Printf("message send successful\n")
		}
	}
}

func main() {
	// parse CLI flags
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [OPTIONS]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
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
	reminder_list := make([]reminder.ReminderV1, 0)
	err = json.Unmarshal(raw_file, &reminder_list)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("parsed reminders: %v\n", reminder_list)

	// main loop
	var wait_time time.Duration = 60
	ticker := time.NewTicker(wait_time*time.Second)
	for {
		received_time := <-ticker.C
		fmt.Printf("firing at time %s\n", received_time)
		fire_reminders(received_time, sns_client, reminder_list)
	}
}
