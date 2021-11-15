package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync/atomic"
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

	errCountPtr *uint32
}

func NewTest(test map[interface{}]interface{}) *Test {
	var zero uint32 = 0
	t := &Test{
		Name:            test["Name"].(string),
		URL:             test["URL"].(string),
		Method:          test["Method"].(string),
		MaxResponseTime: test["MaxResponseTime"].(int),
		StatusCode:      test["StatusCode"].(int),
		errCountPtr:     &zero,
	}
	if test["HeaderRegexps"] != nil {
		t.HeaderRegexps = make(map[string]string)
		for h, rxp := range test["HeaderRegexps"].(map[interface{}]interface{}) {
			t.HeaderRegexps[h.(string)] = rxp.(string)
		}
	}
	if test["ContentRegexp"] != nil {
		t.ContentRegexp = test["ContentRegexp"].(string)
	}
	return t
}

func (t *Test) HighErrorCount() bool {
	return atomic.LoadUint32(t.errCountPtr) >= 3
}

func (t *Test) Run() (err error) {
	defer func() {
		if err != nil {
			done := false
			var val uint32
			for !done {
				val = atomic.LoadUint32(t.errCountPtr)
				done = atomic.CompareAndSwapUint32(t.errCountPtr, val, val+1)
			}
		} else {
			atomic.StoreUint32(t.errCountPtr, 0)
		}
	}()

	req, err := http.NewRequest(t.Method, t.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-store, max-age=0")
	req.Header.Set("User-Agent", "uptime-mon bot/"+version+" (+https://github.com/CHTJonas/uptime-mon)")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Duration(t.MaxResponseTime) * time.Millisecond,
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