#!/bin/bash

set -euo pipefail

export CLICOLOR=1

termshot --clip-canvas --filename .doc/example-cmd-figlet.png --embed README.md --id example-figlet -- lolcat -f <(figlet -f big termshot)
termshot --columns 128 --clip-canvas --filename .doc/example-ls-a.png -- ls -a
termshot --columns 128 --clip-canvas --show-cmd --filename .doc/example-cmd-ls-a.png --embed README.md --id example-ls-a -- ls -a
termshot --columns 128 --clip-canvas --show-cmd --filename .doc/example-cmd-ls-a.png --embed README.md --id example-show-cmd -- ls -a
termshot --columns 128 --clip-canvas --show-cmd --filename .doc/example-cmd-ls-pipe-grep.png --embed README.md --id example-ls-pipe-grep -- "ls -1 | grep go"
