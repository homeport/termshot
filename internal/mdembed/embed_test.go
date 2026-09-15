// Copyright © 2026 The Homeport Team
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

package mdembed_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/homeport/termshot/internal/mdembed"
)

var _ = Describe("Updating markdown files", func() {
	Context("replacing an existing screenshot reference", func() {
		It("should replace the image that follows the marker", func() {
			input := "# docs\n\n<!-- termshot:id=ls-a -->\n![old](old.png)\n\nmore\n"

			updated, err := Apply(input, "ls-a", ".doc/ls-a.png", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal("# docs\n\n<!-- termshot:id=ls-a -->\n![old](.doc/ls-a.png)\n\nmore\n"))
		})

		It("should keep other markers unchanged", func() {
			input := "<!-- termshot:id=one -->\n![a](a.png)\n<!-- termshot:id=two -->\n![b](b.png)\n"

			updated, err := Apply(input, "two", "new.png", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal("<!-- termshot:id=one -->\n![a](a.png)\n<!-- termshot:id=two -->\n![b](new.png)\n"))
		})

		It("should use the provided alt text", func() {
			input := "<!-- termshot:id=demo -->\n![old](old.png)\n"

			updated, err := Apply(input, "demo", "new.png", "fresh")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal("<!-- termshot:id=demo -->\n![fresh](new.png)\n"))
		})
	})

	Context("inserting a screenshot reference", func() {
		It("should insert an image when the marker has no image yet", func() {
			input := "intro\n<!-- termshot:id=figlet -->\n\n## Next\n"

			updated, err := Apply(input, "figlet", "shot.png", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal("intro\n<!-- termshot:id=figlet -->\n![figlet](shot.png)\n\n## Next\n"))
		})

		It("should replace a closed marker region", func() {
			input := "<!-- termshot:id=demo -->\nstale text\n![kept](old.png)\n<!-- /termshot -->\n"

			updated, err := Apply(input, "demo", "new.png", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal("<!-- termshot:id=demo -->\n![kept](new.png)\n<!-- /termshot -->\n"))
		})
	})

	Context("rejecting invalid embed targets", func() {
		It("should fail when the marker is missing", func() {
			_, err := Apply("# README\n", "missing", "out.png", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no embed marker <!-- termshot:id=missing --> found"))
		})

		It("should fail when the marker is duplicated", func() {
			input := "<!-- termshot:id=demo -->\n<!-- termshot:id=demo -->\n"

			_, err := Apply(input, "demo", "out.png", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate embed marker"))
		})

		It("should fail when the id is empty", func() {
			_, err := Apply("<!-- termshot:id=demo -->\n", "", "out.png", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--id is required"))
		})
	})

	Context("writing the markdown file", func() {
		It("should write a path relative to the markdown file", func() {
			dir := GinkgoT().TempDir()
			docs := filepath.Join(dir, "docs")
			Expect(os.Mkdir(docs, os.FileMode(0755))).To(Succeed())

			markdown := filepath.Join(docs, "README.md")
			Expect(os.WriteFile(markdown, []byte("<!-- termshot:id=demo -->\n"), os.FileMode(0644))).To(Succeed())

			image := filepath.Join(docs, "shots", "demo.png")
			Expect(UpdateFile(markdown, "demo", image, "demo")).To(Succeed())

			content, err := os.ReadFile(markdown)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal("<!-- termshot:id=demo -->\n![demo](shots/demo.png)\n"))
		})
	})
})
