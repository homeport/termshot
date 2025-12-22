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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/homeport/termshot/internal/img"
)

var _ = Describe("Creating images", func() {
	Context("Use scaffold to create PNG file", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should write a PNG stream based on provided input", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-foobar.png")))
		})

		It("should omit the window decorations when configured", func() {
			scaffold := NewImageCreator()
			scaffold.DrawDecorations(false)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-no-decoration.png")))
		})

		It("should omit the window shadow when configured", func() {
			scaffold := NewImageCreator()
			scaffold.DrawShadow(false)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-no-shadow.png")))
		})

		It("should clip the canvas when configured", func() {
			scaffold := NewImageCreator()
			scaffold.ClipCanvas(true)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-clip-canvas.png")))
		})

		It("should wrap the content when configured", func() {
			scaffold := NewImageCreator()
			scaffold.SetColumns(4)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-wrapping.png")))
		})

		It("should show the command when configured", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddCommand("echo", "foobar")).To(Succeed())
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).To(Succeed())
			Expect(scaffold).To(LookLike(testdata("expected-show-cmd.png")))
		})

		It("should apply margin correctly", func() {
			scaffold := NewImageCreator()
			scaffold.SetMargin(24)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-margin.png")))
		})

		It("should apply padding correctly", func() {
			scaffold := NewImageCreator()
			scaffold.SetPadding(60)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-padding.png")))
		})

		It("should write a PNG stream based on provided input with ANSI sequences", func() {
			var buf bytes.Buffer
			_, _ = Fprintf(&buf, "Text with emphasis, like *bold*, _italic_, _*bold/italic*_ or ~underline~.\n\n")
			_, _ = Fprintf(&buf, "Colors:\n")
			_, _ = Fprintf(&buf, "\tRed{Red}\n")
			_, _ = Fprintf(&buf, "\tGreen{Green}\n")
			_, _ = Fprintf(&buf, "\tBlue{Blue}\n")
			_, _ = Fprintf(&buf, "\tMintCream{MintCream}\n")

			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(&buf)).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-ansi.png")))
		})
	})

	Context("Use scaffold to create raw output file", func() {
		var buf bytes.Buffer

		BeforeEach(func() {
			SetColorSettings(ON, ON)
			buf.Reset()
		})

		It("should write an output file with the content as-is", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader(Sprintf("MintCream{foobar}")))).To(Succeed())
			Expect(scaffold.WriteRaw(&buf)).To(Succeed())
			Expect(buf.String()).To(Equal("\x1b[38;2;245;255;250mfoobar\x1b[0m"))
		})
	})
})
