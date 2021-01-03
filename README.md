# text-me-when

`text-me-when` sends you SMS reminder messages at times you specify in a `cron`-like
format. It uses AWS SNS to send SMS messages. It is independent of of any phone, email
or calendar ecosystems, and thus protects your privacy.


## Building and Installation

To build `text-me-when`, you need [the Go compiler](https://golang.org/doc/install).
Once you have that, clone this repository and do:

```
go build -o text-me-when main.go
```

You can then copy `text-me-when` to an appropriate location.


## Configuration

### Reminders

Reminders are configured via a JSON file that contains an array of `Reminder` objects.
Each `Reminder` has a message that is sent when it is triggered, and a list
of triggers. Each trigger specifies times when the message is sent out.
Each trigger uses a cron-like format, and supports the formats "*", "*/3", "1,2,3", and
"2". These formats should be familiar to anyone who has used cron; those who
are not familiar with cron should read its documentation to understand how to
configure `text-me-when`.

The following is an example config:

```
[
  {
    "version": "v1",
    "message": "This message is printed on the first day of January and August at 09:00.",
    "triggers": [
      {
        "trigger_type": "cron",
        "minute": "0",
        "hour": "9",
        "day_of_month": "1",
        "month": "1,8",
        "day_of_week": "*"
      }
    ]
  },
  {
    "version": "v1",
    "message": "This message is printed every other minute on the third day of every month.",
    "triggers": [
      {
        "trigger_type": "cron",
        "minute": "*/2",
        "hour": "*",
        "day_of_month": "3",
        "month": "*",
        "day_of_week": "*"
      }
    ]
  }
]
```

The default location for this file is `/etc/text-me-when.json`.
You can change this with the `-c` flag.


### General Config

Other than reminders, there are four pieces of information you need to pass
to `text-me-when`. These are more or less explained in the usage:

```
Usage: text-me-when [OPTIONS] PHONE_NUMBER

  Checks once a minute for reminders whose messages should be sent out.
  PHONE_NUMBER is the phone number, in E.164 format, that you want the messages
  to be sent to.

  The environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and
  AWS_DEFAULT_REGION are required to send text messages via AWS SNS. For more
  information on what these mean please see the AWS documentation.

Options:
  -c string
        The path to the reminders config (default "/etc/text-me-when.json")
  -t    Send a test SMS to the configured phone number before entering main loop
```
