// Copyright © 2025 The Homeport Team
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

package ptexec_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/homeport/termshot/internal/ptexec"
)

var _ = Describe("Pseudo Terminal Execute Suite", func() {
	Context("running commands in pseudo terminal", func() {
		It("should run a command just fine", func() {
			out, err := New().Stdout(GinkgoWriter).
				Command("echo", "hello").
				Run()

			Expect(err).ToNot(HaveOccurred())
			Expect(trimmed(out)).To(Equal("hello"))
		})

		It("should run a script just fine", func() {
			out, err := New().Stdout(GinkgoWriter).
				Command("echo hello").
				Run()

			Expect(err).ToNot(HaveOccurred())
			Expect(trimmed(out)).To(Equal("hello"))
		})

		It("should run with fixed terminal size", func() {
			out, err := New().Stdout(GinkgoWriter).Cols(40).Rows(12).Command("stty", "size").Run()
			Expect(err).ToNot(HaveOccurred())
			Expect(trimmed(out)).To(Equal("12 40"))
		})

		It("should not truncate the output", func() {
			out, err := New().Stdout(GinkgoWriter).
				Command("for c in {a..g}; do echo -n $c; sleep 0.01; done").
				Run()

			Expect(err).ToNot(HaveOccurred())
			Expect(trimmed(out)).To(Equal("abcdefg"))
		})
	})
})
