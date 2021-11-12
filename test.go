package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

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
	req.Header.Set("User-Agent", "uptime-mon bot/"+version+" (+https://github.com/CHTJonas/uptime-mon)")

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
