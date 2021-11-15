package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var version = "dev"
var testArr []*Test

func notificationHelper(format string, a ...interface{}) {
	err := notifyf(format, a...)
	if err != nil {
		fmt.Println("error sending Slack notification:", err)
	}
}

func main() {
	readConfig()
	initTestArr()

	offset := float64(15) / float64(len(testArr))
	duration := time.Duration(offset * float64(time.Second))
	for _, test := range testArr {
		go func(t *Test) {
			for range time.Tick(15 * time.Second) {
				err := t.Run()
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						err = fmt.Errorf("response time was greater than %d milliseconds", t.MaxResponseTime)
					}
					errStr := fmt.Sprintf("Test failed: %s: %s", t.Name, err)
					if version == "dev" {
						fmt.Println(errStr)
					}
					if t.HighErrorCount() {
						notificationHelper(errStr)
					}
				} else if version == "dev" {
					fmt.Println("Test success:", t.Name)
				}
			}
		}(test)
		time.Sleep(duration)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
}

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/uptime-mon/")
	viper.AddConfigPath("$HOME/.config/uptime-mon/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Failed to read config file:", err)
		os.Exit(125)
	}
}

func initTestArr() {
	tests := viper.Get("tests").([]interface{})
	testArr = make([]*Test, len(tests))
	for i, t := range tests {
		testArr[i] = NewTest(t.(map[interface{}]interface{}))
	}
	fmt.Println("Found", len(tests), "tests in config file")
}
