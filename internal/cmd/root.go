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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"
	"github.com/homeport/termshot/internal/img"
	"github.com/homeport/termshot/internal/ptexec"
	"github.com/spf13/cobra"
)

// version string will be injected by automation
var version string

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [flags] [--] command [command flags] [command arguments] [...]", executableName()),
	Short: "Creates a screenshot of terminal command output",
	Long: `Executes the provided command as-is with all flags and arguments in a pseudo
terminal and captures the generated output. The result is printed as it was
produced. Additionally, an image will be rendered in a lookalike terminal
window including all terminal colors and text decorations.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion, err := cmd.Flags().GetBool("version"); showVersion && err == nil {
			if len(version) == 0 {
				version = "(development)"
			}

			bunt.Printf("Lime{*%s*} version DimGray{%s}\n",
				executableName(),
				version,
			)

			return nil
		}

		if len(args) == 0 {
			return cmd.Usage()
		}

		scaffold := img.NewImageCreator()

		// Prepend command line arguments to output content
		if includeCommand, err := cmd.Flags().GetBool("include-command"); err == nil && includeCommand {
			scaffold.AddContent(bunt.Sprintf(
				"Lime{➜} DimGray{%s}\n",
				strings.Join(args, " ")),
			)
		}

		bytes, err := ptexec.RunCommandInPseudoTerminal(args[0], args[1:]...)
		if err != nil {
			return err
		}

		// Allow manual override of command output content
		if edit, err := cmd.Flags().GetBool("edit"); err == nil && edit {
			tmpFile, err := ioutil.TempFile("", executableName())
			if err != nil {
				return err
			}

			defer os.Remove(tmpFile.Name())

			ioutil.WriteFile(tmpFile.Name(), bytes, os.FileMode(0644))

			editor := os.Getenv("EDITOR")
			if len(editor) == 0 {
				editor = "vi"
			}

			if _, err := ptexec.RunCommandInPseudoTerminal(editor, tmpFile.Name()); err != nil {
				return err
			}

			bytes, err = ioutil.ReadFile(tmpFile.Name())
			if err != nil {
				return err
			}
		}

		scaffold.AddContent(string(bytes))

		return scaffold.SavePNG("out.png")
	},
}

// Execute is the main entry point into the CLI code
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		neat.PrintError(err)
		os.Exit(1)
	}
}

func executableName() string {
	if executable, err := os.Executable(); err == nil {
		return filepath.Clean(filepath.Base(executable))
	}

	return "termshot"
}

func init() {
	rootCmd.PersistentFlags().BoolP("version", "v", false, "show version")
	rootCmd.PersistentFlags().BoolP("include-command", "i", false, "include command in screenshot")
	rootCmd.PersistentFlags().BoolP("edit", "e", false, "use system default editor to change content before the screenshot")
}
