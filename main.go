package main

import (
	"time"
	"fmt"
	"flag"
	"io/ioutil"
	"encoding/json"
	"os"
)

type Config struct {
	ConfigPath string
}

type Reminder struct {
	FirstName string
	LastName string
	Birthday string
	//Birthday time.Time
}

//type (r *Reminder) UnmarshalJson(data []byte) error {
//	if string(data) == "null" { return nil }
//	for 
//}

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
