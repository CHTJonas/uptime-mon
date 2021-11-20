# Uptime Monitor

uptime-mon is a small application written in Go that contacts your HTTP services and ensures they're working correctly. It sends a request and checks that the response matches the criteria you set based on a combination of request time, status code, headers and body content. Tests can be made over IPv4, IPv6 or both, and alerting is via Slack.

## Usage

uptime-mon loads its configuration settings by looking for a `config.yml` file in either `/etc/uptime-mon/`, `$HOME/.config/uptime-mon/` or the current working directory in that order of precedence. An example of a config file is included below:

```yaml
settings:
  slack-webhook: https://hooks.slack.com/services/foo/bar

tests:
  - name: Google homepage
    url: https://www.google.co.uk/
    method: GET
    max-response-time: 3000
    status-code: 200
    header-regexps:
      Content-Type: text\/html; charset=ISO-8859-1
      Set-Cookie: .*
    content-regexp: Google Search
    network: tcp4
```

The contents of headers are matched based on standard regular expressions, ditto the response body. Note that sometimes you may need to use quotation marks to force the YAML parser to interpret a particular value as a string. The `network` field allows you to specify whether you want to run the test over IPv4 (`tcp4`), IPv6 (`tcp6`), both (`both`) or use [Happy Eyeballs](https://en.wikipedia.org/wiki/Happy_Eyeballs) (leave blank).

## Installation

Pre-built binaries for a variety of operating systems and architectures are available to download from [GitHub Releases](https://github.com/CHTJonas/uptime-mon/releases). If you wish to compile from source then you will need a suitable [Go toolchain installed](https://golang.org/doc/install). After that just clone the project using Git and run Make! Cross-compilation is easy in Go so by default we build for all targets and place the resulting executables in `./bin`:

```bash
git clone https://github.com/CHTJonas/uptime-mon.git
cd uptime-mon
make clean && make all
```

## Copyright

uptime-mon is licensed under the [BSD 2-Clause License](https://opensource.org/licenses/BSD-2-Clause).

Copyright (c) 2021 Charlie Jonas.
