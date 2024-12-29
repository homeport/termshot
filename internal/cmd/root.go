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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"
	"github.com/gonvenience/wrap"

	"github.com/homeport/termshot/internal/img"
	"github.com/homeport/termshot/internal/ptexec"

	"github.com/spf13/cobra"
)

// version string will be injected by automation
var version string

// saveToClipboard function will be implemented by OS specific code
var saveToClipboard func(img.Scaffold) error

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [%s flags] [--] command [command flags] [command arguments] [...]", executableName(), executableName()),
	Short: "Creates a screenshot of terminal command output",
	Long: `Executes the provided command as-is with all flags and arguments in a pseudo
terminal and captures the generated output. The result is printed as it was
produced. Additionally, an image will be rendered in a lookalike terminal
window including all terminal colors and text decorations.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
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

		var scaffold = img.NewImageCreator()
		var buf bytes.Buffer

		// Initialise scaffold with a column sizing so that the
		// content can be wrapped accordingly
		//
		if columns, err := cmd.Flags().GetInt("columns"); err == nil && columns > 0 {
			scaffold.SetColumns(columns)
		}

		// Disable window shadow if requested
		//
		if val, err := cmd.Flags().GetBool("no-shadow"); err == nil {
			scaffold.DrawShadow(!val)
		}

		// Disable window decorations (buttons) if requested
		//
		if val, err := cmd.Flags().GetBool("no-decoration"); err == nil {
			scaffold.DrawDecorations(!val)
		}

		// Prepend command line arguments to output content
		//
		if includeCommand, err := cmd.Flags().GetBool("show-cmd"); err == nil && includeCommand {
			bunt.Fprintf(&buf, "Lime{➜} DimGray{%s}\n", strings.Join(args, " "))
		}

		bytes, err := ptexec.RunCommandInPseudoTerminal(args[0], args[1:]...)
		if err != nil {
			return err
		}

		buf.Write(bytes)

		// Allow manual override of command output content
		//
		if edit, err := cmd.Flags().GetBool("edit"); err == nil && edit {
			tmpFile, tmpErr := os.CreateTemp("", executableName())
			if tmpErr != nil {
				return tmpErr
			}

			defer os.Remove(tmpFile.Name())

			if err := os.WriteFile(tmpFile.Name(), buf.Bytes(), os.FileMode(0644)); err != nil {
				return err
			}

			editor := os.Getenv("EDITOR")
			if len(editor) == 0 {
				editor = "vi"
			}

			if _, err := ptexec.RunCommandInPseudoTerminal(editor, tmpFile.Name()); err != nil {
				return err
			}

			bytes, tmpErr := os.ReadFile(tmpFile.Name())
			if tmpErr != nil {
				return tmpErr
			}

			buf.Reset()
			buf.Write(bytes)
		}

		if err := scaffold.AddContent(&buf); err != nil {
			return err
		}

		// save image to clipboard
		//
		if toClipboard, err := cmd.Flags().GetBool("clipboard"); err == nil && toClipboard {
			return saveToClipboard(scaffold)
		}

		// save image to file
		//
		filename, err := cmd.Flags().GetString("filename")
		if filename == "" || err != nil {
			fmt.Fprintf(os.Stderr, "failed to read filename from command-line, defaulting to out.png")
			filename = "out.png"
		}

		if extension := filepath.Ext(filename); extension != ".png" {
			return fmt.Errorf("file extension %q of filename %q is not supported, only png is supported", extension, filename)
		}

		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		defer file.Close()
		return scaffold.Write(file)
	},
}

// Execute is the main entry point into the CLI code
func Execute() {
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, e error) error {
		return fmt.Errorf("unknown %s flag %w",
			executableName(),
			fmt.Errorf("issue with %v\n\nIn order to differentiate between program flags and command flags,\nuse '--' before the command so that all flags before the separator\nbelong to %s, while all others are used for the command.\n\n%s", e, executableName(), c.UsageString()),
		)
	})

	if err := rootCmd.Execute(); err != nil {
		var headline string
		var content bytes.Buffer

		switch e := err.(type) {
		case wrap.ContextError:
			headline = e.Context()
			content.WriteString(e.Cause().Error())

		default:
			headline = "Error occurred"
			content.WriteString(e.Error())
		}

		neat.Box(
			os.Stderr,
			headline,
			&content,
			neat.HeadlineColor(bunt.OrangeRed),
			neat.ContentColor(bunt.LightCoral),
			neat.NoLineWrap(),
		)

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
	rootCmd.Flags().SortFlags = false

	// flags to control look
	rootCmd.Flags().BoolP("edit", "e", false, "edit content before creating screenshot")
	rootCmd.Flags().BoolP("show-cmd", "c", false, "include command in screenshot")
	rootCmd.Flags().IntP("columns", "C", 0, "force fixed number of columns in screenshot")
	rootCmd.Flags().Bool("no-decoration", false, "do not draw window decorations")
	rootCmd.Flags().Bool("no-shadow", false, "do not draw window shadow")

	// flags for output related settings
	rootCmd.Flags().StringP("filename", "f", "out.png", "filename of the screenshot")

	// internals
	rootCmd.Flags().BoolP("version", "v", false, "show version")
}
