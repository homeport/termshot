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
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
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

	// Set RAW mode for Stdin
	if isTerminal(os.Stdin) {
		oldState, rawErr := term.MakeRaw(int(os.Stdin.Fd()))
		if rawErr != nil {
			return nil, wrap.Errorf(rawErr, "failed to enable RAW mode for Stdin")
		}

		// And make sure to restore the original mode eventually
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()
	}

	pt, err := pty.Start(exec.Command(name, args...))
	if err != nil {
		return nil, err
	}

	// Support terminal resizing
	if isTerminal(os.Stdin) {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if ptyErr := pty.InheritSize(os.Stdin, pt); ptyErr != nil {
					errors = append(errors, wrap.Error(ptyErr, "error resizing PTY"))
				}
			}
		}()

		ch <- syscall.SIGWINCH
		defer func() {
			signal.Stop(ch)
			close(ch)
		}()
	}

	go func() {
		defer pt.Close()
		_, copyErr := io.Copy(pt, os.Stdin)
		if copyErr != nil {
			errors = append(errors, copyErr)
		}
	}()

	var buf bytes.Buffer
	if err = copy(io.MultiWriter(os.Stdout, &buf), pt); err != nil {
		return nil, err
	}

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "issues in background tasks:\n")
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "- %v\n", err.Error())
		}
	}

	return buf.Bytes(), nil
}

func copy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	if err != nil {
		switch terr := err.(type) {
		case *os.PathError:
			// Workaround for issue https://github.com/creack/pty/issues/100
			// where on Linux systems it can happen that the pseudo terminal
			// process finishes while termshot is trying to read. Assuming
			// that the content is already read, this error is treated the
			// same as if it would be an EOF.
			if terr.Op == "read" && terr.Path == "/dev/ptmx" {
				return nil
			}
		}
	}

	return err
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) ||
		isatty.IsCygwinTerminal(f.Fd())
}
