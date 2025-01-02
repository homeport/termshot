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
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// PseudoTerminal defines the setup for a command to be run in a pseudo
// terminal, e.g. terminal size, or output settings
type PseudoTerminal struct {
	name string
	args []string

	shell string

	cols   uint16
	rows   uint16
	resize bool

	stdout io.Writer
}

// New creates a new pseudo terminal builder
func New() *PseudoTerminal {
	return &PseudoTerminal{
		shell:  "/bin/sh",
		resize: true,
		stdout: os.Stdout,
	}
}

// Cols sets the width/columns for the pseudo terminal
func (c *PseudoTerminal) Cols(cols uint16) *PseudoTerminal {
	c.cols = cols
	return c
}

// Rows sets the lines/rows for the pseudo terminal
func (c *PseudoTerminal) Rows(rows uint16) *PseudoTerminal {
	c.rows = rows
	return c
}

// Stdout sets the writer to be used for the standard output
func (c *PseudoTerminal) Stdout(stdout io.Writer) *PseudoTerminal {
	c.stdout = stdout
	return c
}

// Command sets the command and arguments to be used
func (c *PseudoTerminal) Command(name string, args ...string) *PseudoTerminal {
	c.name = name
	c.args = args
	return c
}

// Run runs the provided command/script with the given arguments in a pseudo
// terminal (PTY) so that the behavior is the same if it would be executed
// in a terminal
func (c *PseudoTerminal) Run() ([]byte, error) {
	if c.name == "" {
		return nil, fmt.Errorf("no command specified")
	}

	// Convenience hack in case command contains a space, for example in case
	// typical construct like "foo | grep" are used.
	if strings.Contains(c.name, " ") {
		c.args = []string{
			"-c",
			strings.Join(append(
				[]string{c.name},
				c.args...,
			), " "),
		}
		c.name = c.shell
	}

	// Set RAW mode for Stdin
	if isTerminal(os.Stdin) {
		oldState, rawErr := term.MakeRaw(int(os.Stdin.Fd()))
		if rawErr != nil {
			return nil, fmt.Errorf("failed to enable RAW mode for Stdin: %w", rawErr)
		}

		// And make sure to restore the original mode eventually
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()
	}

	// collect all errors along the way
	var errors = []error{}

	// #nosec G204 -- since this is exactly what we want, arbitrary commands
	pt, err := c.pseudoTerminal(exec.Command(c.name, c.args...))
	if err != nil {
		return nil, err
	}

	// Support terminal resizing
	if c.resize && isTerminal(os.Stdin) {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if ptyErr := pty.InheritSize(os.Stdin, pt); ptyErr != nil {
					errors = append(errors, fmt.Errorf("error resizing PTY: %w", ptyErr))
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
	if err = copy(io.MultiWriter(c.stdout, &buf), pt); err != nil {
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

func (c *PseudoTerminal) pseudoTerminal(cmd *exec.Cmd) (*os.File, error) {
	if c.cols == 0 && c.rows == 0 {
		return pty.Start(cmd)
	}

	size, err := pty.GetsizeFull(os.Stdout)
	if err != nil {
		// Obtaining terminal size is prone to error in CI systems, e.g. in
		// GitHub Action setup or similar, so only fail if CI is not set
		if !isCI() {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}

		// For CI systems, assume a reasonable default even if the terminal
		// size cannot be obtained through ioctl
		size = &pty.Winsize{Rows: 25, Cols: 80}
	}

	// Overwrite rows if fixed value is configured
	if c.rows != 0 {
		size.Rows = c.rows
	}

	// Overwrite columns if fixed value is configured
	if c.cols != 0 {
		size.Cols = c.cols
	}

	// With fixed rows/cols, terminal resizing support is not useful
	c.resize = false

	return pty.StartWithSize(cmd, size)
}

func copy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	if err != nil {
		switch terr := err.(type) { //nolint:gocritic
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

func isCI() bool {
	ci, ok := os.LookupEnv("CI")
	return ok && ci == "true"
}
