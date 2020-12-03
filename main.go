package main

import (
	"time"
	"fmt"
	"flag"
	"io/ioutil"
	"encoding/json"
	"os"
	"errors"
)

type Config struct {
	ConfigPath string
}

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

func get_config() Config {
	c := Config{}
	config_path := flag.String("c", "/etc/reminder-boi.json", "The path to the config file")
	flag.Parse()

	fmt.Println(*config_path)

	c.ConfigPath = *config_path
	return c
}

func send_message(number string, message string) error {
	fmt.Printf("sending message \"%s\" to %s", message, number)
	return nil
}

func main() {
	// parse CLI flags
	config := get_config()

	// parse config file
	raw_file, err := ioutil.ReadFile(config.ConfigPath)
	if err != nil {
		fmt.Printf("ioutil.ReadFile: %s\n", err)
		os.Exit(1)
	}
	reminder_list := make([]Reminder, 20)
	err = json.Unmarshal(raw_file, &reminder_list)
	if err != nil {
		fmt.Printf("json.Unmarshal: %s\n", err)
		os.Exit(1)
	}

	// main loop
	var wait_time time.Duration = 5
	ticker := time.NewTicker(wait_time*time.Second)
	for {
		time := <-ticker.C

		fmt.Println(time)
		fmt.Println(reminder_list[0])
	}
}
