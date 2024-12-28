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
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/homeport/termshot/internal/img"
)

var _ = Describe("Creating images", func() {
	Context("Use scaffold to create PNG file", func() {
		It("should write a PNG stream based on provided input", func() {
			scaffold := NewImageCreator()

			err := scaffold.AddContent(strings.NewReader("foobar"))
			Expect(err).ToNot(HaveOccurred())

			err = scaffold.Write(io.Discard)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should write a PNG stream based on provided input with ANSI sequences", func() {
			var buf bytes.Buffer
			_, _ = Fprintf(&buf, "Text with emphasis, like *bold*, _italic_, _*bold/italic*_ or ~underline~.\n\n")
			_, _ = Fprintf(&buf, "Colors:\n")
			_, _ = Fprintf(&buf, "\tRed{Red}\n")
			_, _ = Fprintf(&buf, "\tGreen{Green}\n")
			_, _ = Fprintf(&buf, "\tBlue{Blue}\n")
			_, _ = Fprintf(&buf, "\tMintCream{MintCream}\n")


			scaffold := NewImageCreator(true, true)

			err := scaffold.AddContent(&buf)
			Expect(err).ToNot(HaveOccurred())

			err = scaffold.Write(io.Discard)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
