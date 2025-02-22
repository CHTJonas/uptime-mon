package main

import (
	"fmt"
	"runtime"
	"strings"
)

func uaString() string {
	return fmt.Sprintf("Uptime Monitor/%s Go/%s (+https://github.com/CHTJonas/uptime-mon)",
		version, strings.TrimPrefix(runtime.Version(), "go"))
}

func debugPrintLn(a ...any) (n int, err error) {
	if version == "dev" {
		return fmt.Println(a...)
	}
	return 0, nil
}
