<h1 align="center">
  bruter
</h1>

<h4 align="center">Active network services bruteforce tool.</h4>

<p align="center">
<a href="https://goreportcard.com/report/github.com/vflame6/bruter" target="_blank"><img src="https://goreportcard.com/badge/github.com/vflame6/bruter"></a>
<a href="https://github.com/vflame6/bruter/issues"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat"></a>
<a href="https://github.com/vflame6/bruter/releases"><img src="https://img.shields.io/github/release/vflame6/bruter"></a>
</p>

---

`bruter` is a network services bruteforce tool. It supports several services and can be improved to support more.

# Features

Available modules:

- `clickhouse`

# Usage

```shell
bruter -h
```

Here is a help menu for the tool:

```yaml
usage: bruter --username=USERNAME --password=PASSWORD [<flags>] <command> [<args> ...]

bruter is a network services bruteforce tool.

Flags:
  -h, --[no-]help          Show context-sensitive help (also try --help-long and
                           --help-man).
  -q, --[no-]quiet         Enable quiet mode, print results only
  -o, --output=""          Filename to write output in raw format
  -d, --delay=0            Delay between requests in milliseconds
  -u, --username=USERNAME  Username or file with usernames
  -p, --password=PASSWORD  Password or file with passwords
      --[no-]version       Show application version.

Commands:
  help [<command>...]
  clickhouse [<flags>] <target>
```

# Installation

`bruter` requires **go1.25** to install successfully.

```shell
go install -v github.com/vflame6/bruter@latest
```
