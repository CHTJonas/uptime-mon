package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var version = "dev"
var tests *atomic.Value

func notificationHelper(format string, a ...interface{}) {
	err := notifyf(format, a...)
	if err != nil {
		fmt.Println("error sending Slack notification:", err)
	}
}

func init() {
	tests = new(atomic.Value)
	tests.Store(loadConfig())
}

func main() {
	go testLoop()

	go func() {
		reload := make(chan os.Signal, 1)
		signal.Notify(reload, syscall.SIGHUP)
		for range reload {
			fmt.Println("Received SIGHUP: reloading config")
			tests.Store(loadConfig())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
}

func testLoop() {
	i := 0
	for {
		t := tests.Load().([]*Test)
		offset := float64(15) / float64(len(t))
		duration := time.Duration(offset * float64(time.Second))
		if i >= len(t) {
			i = 0
		}
		go func(t *Test) {
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
		}(t[i])
		time.Sleep(duration)
		i++
	}
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

func loadConfig() []*Test {
	readConfig()
	testsInConfig := viper.Get("tests").([]interface{})
	size := len(testsInConfig)
	tests := make([]*Test, size)
	for i, t := range testsInConfig {
		tests[i] = NewTest(t.(map[interface{}]interface{}))
	}
	fmt.Println("Found", size, "tests in config file")
	return tests
}
