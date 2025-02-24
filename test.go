package main

import (
	"context"
	"fmt"
	"io"
	"net"
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
	NotifyErrCount  int
	HeaderRegexps   map[string]string
	ContentRegexp   string
	Network         string

	errCountPtrMap map[string]*uint32
	notified       bool
}

func NewTest(test map[interface{}]interface{}) *Test {
	t := &Test{
		Name:            test["name"].(string),
		URL:             test["url"].(string),
		Method:          test["method"].(string),
		MaxResponseTime: test["max-response-time"].(int),
		StatusCode:      test["status-code"].(int),
		NotifyErrCount:  test["notify-error-count"].(int),
		errCountPtrMap:  make(map[string]*uint32),
	}
	if test["header-regexps"] != nil {
		t.HeaderRegexps = make(map[string]string)
		for h, rxp := range test["header-regexps"].(map[interface{}]interface{}) {
			t.HeaderRegexps[h.(string)] = rxp.(string)
		}
	}
	if test["content-regexp"] != nil {
		t.ContentRegexp = test["content-regexp"].(string)
	}
	if test["network"] != nil {
		t.Network = test["network"].(string)
	}
	return t
}

func (t *Test) ShouldNotify() bool {
	alertThreshold := uint32(t.NotifyErrCount)
	for errType, errCountPtr := range t.errCountPtrMap {
		currentErrorCount := atomic.LoadUint32(errCountPtr)
		if currentErrorCount >= alertThreshold {
			debugPrintLn(errType, "error counter", currentErrorCount, "for", t.Name, "is above alert thresholf of", alertThreshold)
			return true
		}
		debugPrintLn(errType, "error counter", currentErrorCount, "for", t.Name, "is below alert thresholf of", alertThreshold)
	}
	return false
}

func (t *Test) Run() (err error) {
	switch t.Network {
	case "both":
		for _, networkOverride := range []string{"tcp4", "tcp6"} {
			if err := t.innerRun(networkOverride); err != nil {
				return err
			}
		}
		return nil
	default:
		return t.innerRun(t.Network)
	}
}

func (t *Test) innerRun(networkOverride string) (err error) {
	if networkOverride != "" {
		debugPrintLn("running test with network override", networkOverride)
	}

	defer func() {
		errCountPtr, exists := t.errCountPtrMap[networkOverride]
		if !exists {
			debugPrintLn("initialising", networkOverride, "error counter for", t.Name)
			var zero uint32 = 0
			t.errCountPtrMap[networkOverride] = &zero
			errCountPtr = &zero
		}
		if err != nil {
			done := false
			var val uint32
			for !done {
				val = atomic.LoadUint32(errCountPtr)
				done = atomic.CompareAndSwapUint32(errCountPtr, val, val+1)
			}
			debugPrintLn("incremented", networkOverride, "error counter for", t.Name, "from", val, "to", val+1)
		} else {
			atomic.StoreUint32(errCountPtr, 0)
			debugPrintLn("set", networkOverride, "error counter for", t.Name, "to 0")
		}
	}()

	req, err := http.NewRequest(t.Method, t.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-store, max-age=0")
	req.Header.Set("User-Agent", uaString())

	client := t.getHTTPClient(networkOverride)
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

func (t *Test) getHTTPClient(networkOverride string) *http.Client {
	dialer := &net.Dialer{
		KeepAlive: -1,
	}
	dialCtx := func(ctx context.Context, network, addr string) (net.Conn, error) {
		if networkOverride != "" {
			network = networkOverride
		}
		return dialer.DialContext(ctx, network, addr)
	}
	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext:       dialCtx,
	}
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Duration(t.MaxResponseTime) * time.Millisecond,
	}
}
