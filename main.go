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

func main() {
	readConfig()
	initTestArr()

	offset := float64(30) / float64(len(testArr))
	duration := time.Duration(offset * float64(time.Second))
	for _, test := range testArr {
		go func(t *Test) {
			for range time.Tick(30 * time.Second) {
				err := t.Run()
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						err = fmt.Errorf("response time was greater than %d milliseconds", t.MaxResponseTime)
					}
					fmt.Printf("Test failed: %s: %s\n", t.Name, err)
				} else if version == "dev" {
					fmt.Printf("Test success: %s\n", t.Name)
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
		test := t.(map[interface{}]interface{})
		testArr[i] = &Test{
			Name:            test["Name"].(string),
			URL:             test["URL"].(string),
			Method:          test["Method"].(string),
			MaxResponseTime: test["MaxResponseTime"].(int),
			StatusCode:      test["StatusCode"].(int),
		}
		if test["HeaderRegexps"] != nil {
			testArr[i].HeaderRegexps = make(map[string]string)
			for h, rxp := range test["HeaderRegexps"].(map[interface{}]interface{}) {
				testArr[i].HeaderRegexps[h.(string)] = rxp.(string)
			}
		}
		if test["ContentRegexp"] != nil {
			testArr[i].ContentRegexp = test["ContentRegexp"].(string)
		}
	}
	fmt.Println("Found", len(tests), "tests in config file")
}
