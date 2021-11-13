package main

import (
	"fmt"
	"os"

	"github.com/nlopes/slack"
)

var slackWebhook string

func init() {
	slackWebhook = os.Getenv("SLACK_WEBHOOK")
}

func notify(msg string) error {
	return slack.PostWebhook(slackWebhook, &slack.WebhookMessage{
		Text: msg,
	})
}

func notifyf(format string, a ...interface{}) error {
	return notify(fmt.Sprintf(format, a...))
}
