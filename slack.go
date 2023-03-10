package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

func notify(msg string) error {
	hookURL := getConfig().slackWebhook
	return slack.PostWebhook(hookURL, &slack.WebhookMessage{
		Text: msg,
	})
}

func notifyf(format string, a ...interface{}) error {
	return notify(fmt.Sprintf(format, a...))
}
