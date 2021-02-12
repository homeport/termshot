// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ptexec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/gonvenience/wrap"
	terminal "golang.org/x/term"
)

// RunCommandInPseudoTerminal runs the provided program with the given
// arguments in a pseudo terminal (PTY) so tha the behavior is the same
// if it would be executed in a terminal
func RunCommandInPseudoTerminal(name string, args ...string) ([]byte, error) {
	var errors = []error{}

	// Convenience hack in case command contains a space, for example in case
	// typical construct like "foo | grep" are used.
	if strings.Contains(name, " ") {
		args = []string{
			"-c",
			strings.Join(append(
				[]string{name},
				args...,
			), " "),
		}
		name = "/bin/sh"
	}

	pt, err := pty.Start(exec.Command(name, args...))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = pt.Close()
	}()

	// Support terminal resizing
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, pt); err != nil {
				errors = append(errors, wrap.Error(err, "error resizing PTY"))
			}
		}
	}()
	ch <- syscall.SIGWINCH

	// Set RAW mode for stdin
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, wrap.Errorf(err, "failed to enable RAW mode for stdin")
	}

	// Make sure to restore the original mode
	defer func() {
		_ = terminal.Restore(int(os.Stdin.Fd()), oldState)
	}()

	go func() {
		if _, err := io.Copy(pt, os.Stdin); err != nil {
			errors = append(errors, err)
		}
	}()

	var buf bytes.Buffer
	_, err = io.Copy(io.MultiWriter(os.Stdout, &buf), pt)
	if err != nil {
		return nil, err
	}

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "issues in backgroup tasks:\n")
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "- %v\n", err.Error())
		}
	}

	return buf.Bytes(), nil
}
