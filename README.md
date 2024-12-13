# termshot

[![License](https://img.shields.io/github/license/homeport/termshot.svg)](https://github.com/homeport/termshot/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/homeport/termshot)](https://goreportcard.com/report/github.com/homeport/termshot)
[![Tests](https://github.com/homeport/termshot/workflows/Tests/badge.svg)](https://github.com/homeport/termshot/actions?query=workflow%3A%22Tests%22)
[![Codecov](https://img.shields.io/codecov/c/github/homeport/termshot/main.svg)](https://codecov.io/gh/homeport/termshot)
[![Go Reference](https://pkg.go.dev/badge/github.com/homeport/termshot.svg)](https://pkg.go.dev/github.com/homeport/termshot)
[![Release](https://img.shields.io/github/release/homeport/termshot.svg)](https://github.com/homeport/termshot/releases/latest)

Generate beautiful screenshots of your terminal, from your terminal.

```sh
termshot --show-cmd -- lolcat -f <(figlet -f big foobar)
```

This command generates this screenshot:

![example](https://github.com/homeport/termshot/assets/3084745/8ef3a713-70d0-49e3-98e4-e340763dc0b8)

## Installation

To install with Homebrew on macOS or Linux:

```sh
brew install homeport/tap/termshot
```

See [Releases](https://github.com/homeport/termshot/releases/) for pre-compiled binaries for Darwin and Linux.

## Usage

This tool reads the console output and renders an output image that resembles a user interface window. It's inspired by some other web-based tools like [carbon.now.sh](https://carbon.now.sh/), and [codekeep.io/screenshot](https://codekeep.io/screenshot). Unlike those tools, `termshot` does not blindly apply syntax highlighting to some provided text; instead it reads the ANSI escape codes ("rich text") logged by most command-line tools and uses it to generate a high-fidelity "screenshot" of your terminal output.

Like `time`, `watch`, or `perf`, just prefix the command you want to screenshot with `termshot -- `.

```sh
termshot -- ls -a
```

This will generate an image file called `out.png` in the current directory.

![basic termshot](https://github.com/homeport/termshot/assets/3084745/11b578ee-8106-4e71-a1b8-57bbca4b192f)

In some cases, if your target command contains _pipes_—there may still be ambiguity, even with `--`. In these cases, wrap your command in double quotes.

```sh
termshot -- "ls -l | grep go"
```

![termshot with pipes](https://github.com/homeport/termshot/assets/3084745/5d0dd1ab-820d-46fc-8af7-8a294193c5ca)

#### `--show-cmd`/`-c`

Include the target command in the screenshot.

```sh
termshot --show-cmd -- "ls -a"
termshot --c -- "ls -a"
```

![termshot that shows command](https://github.com/homeport/termshot/assets/3084745/3fbdd952-785d-4865-b216-f33bdaceb4da)

#### `--edit`/`-e`

Edit the output before generating the screenshot. This will open the rich text output in the editor configured in `$EDITOR`, using `vi` as a fallback. Use this flag to remove unwanted or sensitive output.

```sh
termshot --edit -- "ls -a"
termshot -e -- "ls -a"
```

#### `--filename`/`-f`

Specify a path where the screenshot should be generated. This can be an absolute path or a relative path; relative paths will be resolved relative to the current working directory.

```sh
termshot -- "ls -a" # defaults to <cwd>/out.png
termshot --filename my-image.png -- "ls -a"
termshot --filename screenshots/my-image.png -- "ls -a"
termshot --filename /Desktop/my-image.png -- "ls -a"
```

Defaults to `out.png`.

#### `--version`/`-v`

Print the version of `termshot` installed.

```sh
$ termshot --version
termshot version 0.2.5
```

![out](https://github.com/homeport/termshot/assets/3084745/3fbdd952-785d-4865-b216-f33bdaceb4da)

### Multiple commands

In order to work, `termshot` uses a pseudo terminal for the command to be executed. For advanced use cases, you can invoke a fully interactive shell, run several commands, and capture the entire output. The screenshot will be created once you terminate the shell.

```sh
termshot /bin/zsh
```

> _Please note:_ This project is work in progress. Although a lot of the ANSI sequences can be parsed, there are definitely commands in existence that create output that cannot be parsed correctly, yet. Also, commands that reset the cursor position are known to create issues.
