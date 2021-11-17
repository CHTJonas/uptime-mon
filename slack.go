package main

import (
	"fmt"

	"github.com/nlopes/slack"
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
