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
