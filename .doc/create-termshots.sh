#!/bin/bash

set -euo pipefail

export CLICOLOR=1

termshot --clip-canvas --filename .doc/example-cmd-figlet.png -- lolcat -f <(figlet -f big termshot)
termshot --columns 128 --clip-canvas --filename .doc/example-ls-a.png -- ls -a
termshot --columns 128 --clip-canvas --show-cmd --filename .doc/example-cmd-ls-a.png -- ls -a
termshot --columns 128 --clip-canvas --show-cmd --filename .doc/example-cmd-ls-pipe-grep.png -- "ls -1 | grep go"
