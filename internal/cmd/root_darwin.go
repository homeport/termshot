// Copyright © 2023 The Homeport Team
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

package cmd

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"

	"github.com/homeport/termshot/internal/img"
)

const osascript = "/usr/bin/osascript"

// hasOsascript checks if /usr/bin/osascript exists and is executable
func hasOsascript() bool {
	if fi, err := os.Stat(osascript); err == nil {
		return fi.Mode()&0111 != 0
	}

	return false
}

func init() {
	if hasOsascript() {
		// register tool flag to enable clipboard option
		rootCmd.Flags().BoolP("clipboard", "b", false, "copy termshot to clipboard, overrules filename option")

		// register function to copy image into the clipboard
		saveToClipboard = func(scaffold img.Scaffold) error {
			var buf bytes.Buffer

			if _, err := buf.WriteString("set the clipboard to «data PICT"); err != nil {
				return err
			}

			if err := scaffold.Write(hex.NewEncoder(&buf)); err != nil {
				return err
			}

			if _, err := buf.WriteString("»"); err != nil {
				return err
			}

			cmd := exec.Command(osascript, "-e", buf.String())
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Fprint(os.Stderr, string(out))
				return err
			}

			return nil
		}
	}
}
