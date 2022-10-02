package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akerl/go-lambda/apigw/events"
	"github.com/akerl/go-lambda/mux"
	"github.com/akerl/go-lambda/s3"
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
	for _, c := range config.Checks() {
		if requestKey == c.Key() {
			return events.Response{}, nil
		}
	}
	return events.Reject("invalid token")
}

func handleCheck(r events.Request) (events.Response, error) {
	params := events.Params{Request: &req}
	bucketName := params.Lookup("bucket")
	requestKeyPath := strings.TrimPrefix(r.Path, "/")

	client := s3.Client()
	stamp := time.Now().Unix()

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &requestKeyPath,
		Body:   strings.NewReader(string(stamp)),
	})

	if err != nil {
		return events.Fail("failed to write object")
	}

	return events.Succeed(fmt.Sprintf("check updated: %d", stamp))
}

func handleScan(r events.Request) (events.Response, error) {
	params := events.Params{Request: &req}
	bucketName := params.Lookup("bucket")

	for _, c := range config.Checks {
		requestKeyPath := "checks/" + c.Key()
		stampBytes, err := s3.GetObject(bucketName, requestKeyPath)
		if err != nil {
			return events.Fail(fmt.Sprintf("failed parsing %s", requestKeyPath))
		}
		stamp, err := strconv.ParseInt(string(stampBytes), 10, 64)
		if err != nil {
			return events.Fail(fmt.Sprintf("failed parsing %s", requestKeyPath))
		}
		last := time.Unix(stamp, 0)
		expiry := last.Add(c.Threshold * time.Minute)
		now := time.Now()
		if expiry.Before(now) {
			msg := slack.WebhookMessage{
				Text: fmt.Sprintf("Expired heartbeats for %s", c.Name()),
			}
			err := slack.PostWebhook(config.SlackWebhook, &msg)
			if err != nil {
				return events.Fail(fmt.Sprintf("failed posting message"))
			}
		}
	}
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
