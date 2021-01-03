package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
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

func fire_reminders(call_time time.Time, phone_number string, sns_client *sns.SNS,
	reminder_list []reminder.ReminderV1) {
	for _, reminder := range reminder_list {
		if reminder.ShouldRun(call_time) {
			err := send_message(sns_client, reminder.Message, phone_number)
			if err != nil {
				log.Printf("send_message failed: %s", err)
				continue
			}
			log.Printf("sent message \"%s\" to %s", reminder.Message, phone_number)
		}
	}
}

func main() {
	// set up logging
	log.SetOutput(os.Stdout)

	// parse CLI flags
	flag.Usage = func() {
		usage_header := "%s [OPTIONS] PHONE_NUMBER\n" +
			"\n" +
			"  Checks once a minute for reminders whose messages should be sent out.\n" +
			"  PHONE_NUMBER is the phone number, in E.164 format, that you want the messages\n" +
			"  to be sent to.\n" +
			"\n" +
			"  The environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and\n" +
			"  AWS_DEFAULT_REGION are required to send text messages via AWS SNS. For more\n" +
			"  information on what these mean please see the AWS documentation.\n" +
			"\n" +
			"Options:\n"
		fmt.Fprintf(flag.CommandLine.Output(), usage_header, os.Args[0])
		flag.PrintDefaults()
	}
	reminders_path := flag.String("c", "/etc/text-me-when.json", "The path to the reminders config")
	send_test := flag.Bool("t", false, "Send a test SMS to the configured phone number before entering main loop")
	flag.Parse()

	// parse phone number
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	phone_number := flag.Args()[0]
	match, err := regexp.MatchString(`^\+[0-9]{11,15}$`, phone_number)
	if err != nil {
		fmt.Printf("There was a problem while validating phone number: %s\n", err)
		os.Exit(1)
	}
	if !match {
		fmt.Printf("%s is not a valid phone number. It must consist of a + followed by up to 15 digits.\n")
		os.Exit(1)
	}

	// read environment variables
	region_key := "AWS_DEFAULT_REGION"
	region, ok := os.LookupEnv(region_key)
	if !ok {
		fmt.Printf("Could not find required env var %s.\n", region_key)
		os.Exit(1)
	}

	// construct sns client
	creds := credentials.NewEnvCredentials()
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(region)
	session := session.Must(session.NewSession(cfg))
	sns_client := sns.New(session)
	log.Print("constructed AWS SNS client")

	// parse config file
	raw_file, err := ioutil.ReadFile(*reminders_path)
	if err != nil {
		fmt.Printf("Failed to read config file: %s\n", err)
		os.Exit(1)
	}
	reminder_list := make([]reminder.ReminderV1, 0)
	err = json.Unmarshal(raw_file, &reminder_list)
	if err != nil {
		fmt.Printf("Failed to parse config file: %s\n", err)
		os.Exit(1)
	}
	log.Printf("read in %d reminders from reminder config", len(reminder_list))

	// send test message if configured
	if *send_test {
		msg := "text-me-when: this is a test message. If you got this, " +
			"you can be sure that message sending is working."
		err := send_message(sns_client, msg, phone_number)
		if err != nil {
			fmt.Printf("There was a problem with sending test message: %s\n", err)
			os.Exit(1)
		}
		log.Printf("sent test message to %s", phone_number)
	}

	// main loop
	var wait_time time.Duration = 60
	log.Print("entering main loop")
	ticker := time.NewTicker(wait_time * time.Second)
	for {
		received_time := <-ticker.C
		log.Print("checking reminders")
		fire_reminders(received_time, phone_number, sns_client, reminder_list)
	}
}
