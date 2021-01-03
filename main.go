package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/adamkpickering/reminder-boi/reminder"
)

func send_message(sns_client *sns.SNS, message string, number string) error {
	pi := &sns.PublishInput{
		Message:     &message,
		PhoneNumber: &number,
	}
	if err := pi.Validate(); err != nil {
		return fmt.Errorf("pi.Validate: %w", err)
	}
	_, err := sns_client.Publish(pi)
	if err != nil {
		return fmt.Errorf("sns_client.Publish: %w", err)
	}
	return nil
}

func fire_reminders(call_time time.Time, sns_client *sns.SNS, reminder_list []reminder.ReminderV1) {
	message := "this is a test message"
	phone_number := "+12345678901"
	for _, reminder := range reminder_list {
		if reminder.ShouldRun(call_time) {
			err := send_message(sns_client, message, phone_number)
			if err != nil {
				log.Printf("send_message failed: %w", err)
				continue
			}
			log.Printf("Sent message \"%s\" to %s", message, phone_number)
		}
	}
}

func main() {
	// set up logging
	log.SetOutput(os.Stdout)

	// parse CLI flags
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [OPTIONS]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	config_path := flag.String("c", "/etc/reminder-boi.json", "The path to the config file")
	flag.Parse()

	// construct sns client
	home_dir, ok := os.LookupEnv("HOME")
	if !ok {
		log.Print("HOME environment variable is not present. Exiting...")
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
		log.Printf("Failed to read config file: %s", err)
		os.Exit(1)
	}
	reminder_list := make([]reminder.ReminderV1, 0)
	err = json.Unmarshal(raw_file, &reminder_list)
	if err != nil {
		log.Printf("Failed to parse config file: %s", err)
		os.Exit(1)
	}

	// main loop
	var wait_time time.Duration = 60
	ticker := time.NewTicker(wait_time * time.Second)
	for {
		received_time := <-ticker.C
		log.Print("checking reminders")
		fire_reminders(received_time, sns_client, reminder_list)
	}
}
