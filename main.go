package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var version = "dev"
var testArr []*Test

type Test struct {
	Name            string
	URL             string
	Method          string
	MaxResponseTime int
	StatusCode      int
	HeaderRegexps   map[string]string
	ContentRegexp   string
}

func (t *Test) Run() error {
	req, err := http.NewRequest(t.Method, t.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-store, max-age=0")
	req.Header.Set("User-Agent", "uptime bot/"+version+" (+https://github.com/CHTJonas/uptime)")

	start := time.Now()
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	if elapsed > time.Duration(t.MaxResponseTime*1000*1000) {
		return fmt.Errorf("response time %d was greater than %d", elapsed, t.MaxResponseTime)
	}
	if resp.StatusCode != t.StatusCode {
		return fmt.Errorf("status code %d did not match %d", resp.StatusCode, t.StatusCode)
	}
	for k, v := range t.HeaderRegexps {
		headers := resp.Header[k]
		if len(headers) == 0 {
			return fmt.Errorf("%s header not present in response", k)
		}
		matched, err := regexp.MatchString(v, headers[0])
		if err != nil {
			return err
		}
		if !matched {
			return fmt.Errorf("%s header did not match %s", k, v)
		}
	}
	matched, err := regexp.MatchString(t.ContentRegexp, string(bodyBytes))
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("response body did not match %s", t.ContentRegexp)
	}
	return nil
}

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
					fmt.Println(t.Name, "test failed:", err)
				} else {
					fmt.Println(t.Name, "test success!")
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
	viper.AddConfigPath("/etc/uptime/")
	viper.AddConfigPath("$HOME/.config/uptime/")
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
