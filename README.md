# termshot

[![License](https://img.shields.io/github/license/homeport/termshot.svg)](https://github.com/homeport/termshot/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/homeport/termshot)](https://goreportcard.com/report/github.com/homeport/termshot)
[![Build Status](https://travis-ci.com/homeport/termshot.svg?branch=main)](https://travis-ci.com/homeport/termshot)
[![Codecov](https://img.shields.io/codecov/c/github/homeport/termshot/main.svg)](https://codecov.io/gh/homeport/termshot)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/homeport/termshot)](https://pkg.go.dev/github.com/homeport/termshot)
[![Release](https://img.shields.io/github/release/homeport/termshot.svg)](https://github.com/homeport/termshot/releases/latest)

The `termshot` tool creates an image that resembles a user interface window similar to what [Carbon](https://carbon.now.sh/), [Instacode](http://instaco.de/), or [codekeep.io/screenshot](https://codekeep.io/screenshot) do for code. However, it works completely offline and takes its input directly from a terminal command output. The ANSI escape codes for Select Graphic Rendition are used for text emphasis (bold, italic, and underline) and of course for terminal colors. Like `time`, `watch`, or `perf`, just place `termshot` before the command.

For example, create a screenshot of the command execution of [`dyff`](https://github.com/homeport/dyff), which relies heavily on terminal colors for text highlighting:

```text
$ termshot dyff between https://raw.githubusercontent.com/cloudfoundry/cf-deployment/v1.10.0/cf-deployment.yml https://raw.githubusercontent.com/cloudfoundry/cf-deployment/v1.20.0/cf-deployment.yml
     _        __  __
   _| |_   _ / _|/ _|  between https://raw.githubusercontent.com/cloudfoundry/cf-deployment/v1.10.0/cf-deployment.yml
 / _' | | | | |_| |_       and https://raw.githubusercontent.com/cloudfoundry/cf-deployment/v1.20.0/cf-deployment.yml
| (_| | |_| |  _|  _|
 \__,_|\__, |_| |_|   returned 80 differences
        |___/

manifest_version
  ± value change
    - v1.10.0
    + v1.20.0

instance_groups.database.jobs.mysql.properties.cf_mysql.mysql.port
  ± value change
    - 33306
    + 13306

[...]
```

This will create a _(fake)_ screenshot which looks like this: ![example](.docs/images/example.png?raw=true "example screenshot")

_Please note:_ This project is work in progress. Although a lot of the ANSI sequences can be parsed, there are definitely commands in existence that create output that cannot be parsed correctly, yet. Also, commands that reset the cursor position are known to create issues.
