# Copyright © 2020 The Homeport Team
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

sources := $(wildcard cmd/*/*.go internal/*/*.go)

.PHONY: clean
clean:
	@rm -rf tmp dist internal/img/font-hack.go
	@go clean -i -cache $(shell go list ./...)

tmp/hack/ttf:
	@mkdir -p tmp/hack
	@/bin/sh -c "echo '\n\033[1mDownloading default font for embedding ...\033[0m'"
	curl --fail --silent --location https://github.com/source-foundry/Hack/releases/download/v3.003/Hack-v3.003-ttf.tar.gz | tar -xzf - -C tmp/hack

internal/img/font-hack.go: tmp/hack/ttf
	@/bin/sh -c "echo '\n\033[1mCreating embedded fonts in Go source ...\033[0m'"
	go-bindata \
	  -pkg img \
	  -nomemcopy \
	  -prefix tmp/hack/ttf \
	  -o internal/img/font-hack.go \
	  tmp/hack/ttf/

.PHONY: test
test: internal/img/font-hack.go $(sources)
	ginkgo \
	  -r \
	  -v \
	  -randomizeAllSpecs \
	  -randomizeSuites \
	  -failOnPending \
	  -nodes=4 \
	  -compilers=2 \
	  -slowSpecThreshold=30 \
	  -race \
	  -trace \
	  -cover
