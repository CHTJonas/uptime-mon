package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var version = "dev"

func main() {
	go testLoop()

	go func() {
		reload := make(chan os.Signal, 1)
		signal.Notify(reload, syscall.SIGHUP)
		for range reload {
			fmt.Println("Received SIGHUP: reloading config")
			config.Store(loadConfig())
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
		c := getConfig()
		offset := float64(15) / float64(len(c.tests))
		duration := time.Duration(offset * float64(time.Second))
		if i >= len(c.tests) {
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
				if t.ShouldNotify() && !t.notified {
					err := notify(errStr)
					if err != nil {
						fmt.Println("error sending Slack notification:", err)
					} else {
						t.notified = true
					}
				}
			} else {
				if version == "dev" {
					fmt.Println("Test success:", t.Name)
				}
				if t.notified {
					notifyf("Test recovered: %s", t.Name)
				}
				t.notified = false
			}
		}(c.tests[i])
		time.Sleep(duration)
		i++
	}
}
