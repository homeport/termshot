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

package img_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/homeport/termshot/internal/img"
)

var _ = Describe("Creating images", func() {
	Context("Use scaffold to create PNG file", func() {
		var withTempFile = func(f func(name string)) {
			SetColorSettings(ON, ON)
			defer SetColorSettings(AUTO, AUTO)

			file, err := ioutil.TempFile("", "termshot.png")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			f(file.Name())
		}

		It("should create a PNG file based on provided input", func() {
			withTempFile(func(name string) {
				scaffold := NewImageCreator()
				scaffold.AddContent(strings.NewReader("foobar"))

				err := scaffold.SavePNG(name)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		It("should create a PNG file based on provided input with ANSI sequences", func() {
			withTempFile(func(name string) {
				var buf bytes.Buffer
				Fprintf(&buf, "Text with emphasis, like *bold*, _italic_, _*bold/italic*_ or ~underline~.\n\n")
				Fprintf(&buf, "Colors:\n")
				Fprintf(&buf, "\tRed{Red}\n")
				Fprintf(&buf, "\tGreen{Green}\n")
				Fprintf(&buf, "\tBlue{Blue}\n")
				Fprintf(&buf, "\tMintCream{MintCream}\n")

				scaffold := NewImageCreator()
				scaffold.AddContent(&buf)

				err := scaffold.SavePNG(name)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
