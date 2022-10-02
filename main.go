package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akerl/go-lambda/apigw/events"
	"github.com/akerl/go-lambda/mux"
	"github.com/akerl/go-lambda/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/slack-go/slack"
)

func isCheckInput(r events.Request) bool {
	return strings.HasPrefix(r.Path, "/checks/")
}

func isScan(r events.Request) bool {
	return r.Path == "/scan"
}

func doesCheckExist(r events.Request) (events.Response, error) {
	requestKey := strings.TrimPrefix(r.Path, "/checks/")
	for _, c := range config.Checks {
		if requestKey == c.Key() {
			return events.Response{}, nil
		}
	}
	return events.Reject("invalid token")
}

func handleCheck(r events.Request) (events.Response, error) {
	params := events.Params{Request: &r}
	bucketName := params.Lookup("bucket")
	requestKeyPath := strings.TrimPrefix(r.Path, "/")
	stamp := time.Now().Unix()

	err := s3.PutObject(bucketName, requestKeyPath, fmt.Sprintf("%d", stamp))
	if err != nil {
		fmt.Printf("%s", err)
		return events.Fail("failed to write object")
	}

	return events.Succeed(fmt.Sprintf("check updated: %d", stamp))
}

func handleScan(r events.Request) (events.Response, error) {
	params := events.Params{Request: &r}
	bucketName := params.Lookup("bucket")

	for _, c := range config.Checks {
		requestKeyPath := "checks/" + c.Key()
		stampBytes, err := s3.GetObject(bucketName, requestKeyPath)
		if err != nil {
			var nsk *types.NoSuchKey
			if errors.As(err, &nsk) {
				stampBytes = []byte("0")
			}
			fmt.Printf("%s", err)
			return events.Fail(fmt.Sprintf("failed parsing %s", requestKeyPath))
		}
		stamp, err := strconv.ParseInt(string(stampBytes), 10, 64)
		if err != nil {
			fmt.Printf("%s", err)
			return events.Fail(fmt.Sprintf("failed converting %s", requestKeyPath))
		}
		last := time.Unix(stamp, 0)
		expiry := last.Add(time.Duration(c.Threshold) * time.Minute)
		now := time.Now()
		if expiry.Before(now) {
			msg := slack.WebhookMessage{
				Text: fmt.Sprintf("Expired heartbeats for %s", c.Name),
			}
			err := slack.PostWebhook(config.SlackWebhook, &msg)
			if err != nil {
				fmt.Printf("%s", err)
				return events.Fail(fmt.Sprintf("failed posting message"))
			}
		}
	}

	return events.Succeed("scanned all checks")
}

func main() {
	var err error
	config, err = loadConfig()
	if err != nil {
		panic(err)
	}

	d := mux.NewDispatcher(
		&mux.SimpleReceiver{
			CheckFunc:  isCheckInput,
			HandleFunc: handleCheck,
			AuthFunc:   doesCheckExist,
		},
		&mux.SimpleReceiver{
			CheckFunc:  isScan,
			HandleFunc: handleScan,
		},
	)
	mux.Start(d)
}
