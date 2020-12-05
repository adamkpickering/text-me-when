package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sns"
)

func main() {
	creds := credentials.NewSharedCredentials("/home/adam/.aws/credentials", "default")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion("ca-central-1")
	session := session.Must(session.NewSession(cfg))
	sns_client := sns.New(session)

	msg := "hello world"
	number := "fake phone number"
	pi := &sns.PublishInput{
		Message: &msg,
		PhoneNumber: &number,
	}
	if err := pi.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "main: %s", err)
		os.Exit(1)
	}
	fmt.Printf("sending msg \"%s\" to \"%s\"\n", msg, number)
	_, err = sns_client.Publish(pi)
	if err != nil {
		fmt.Fprintf(os.Stderr, "main: %s", err)
		os.Exit(1)
	}
}
