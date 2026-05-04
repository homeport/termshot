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

package img

import (
	"os"
	"path/filepath"

	"github.com/golang/freetype/truetype"
	imgfont "golang.org/x/image/font"
)

// indexedFace wraps a truetype font.Face and its underlying *truetype.Font so
// that hasGlyph can use Font.Index(r) != 0 — the correct glyph-presence test
// — rather than the face's ok return, which is true even for .notdef glyphs.
type indexedFace struct {
	imgfont.Face
	f *truetype.Font
}

func (t *indexedFace) hasGlyph(r rune) bool {
	return t.f.Index(r) != 0
}

// loadHackPrimary returns an *indexedFace for the given Hack TTF file found in
// ~/Library/Fonts, enabling correct glyph-presence detection for fallback
// dispatch. Falls back to the provided face if the file is absent or invalid.
func loadHackPrimary(filename string, opts *truetype.Options, fallback imgfont.Face) imgfont.Face {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, "Library/Fonts", filename))
	if err != nil {
		return fallback
	}
	f, err := truetype.Parse(data)
	if err != nil {
		return fallback
	}
	return &indexedFace{Face: truetype.NewFace(f, opts), f: f}
}

var systemFontPaths = []string{
	"/System/Library/Fonts/Apple Symbols.ttf",
	"/System/Library/Fonts/Apple Braille.ttf",
}

// loadSystemFallbacks loads supplementary system fonts as indexed fallback
// faces. Silently skips any that cannot be read or parsed.
func loadSystemFallbacks(opts *truetype.Options) []imgfont.Face {
	home, _ := os.UserHomeDir()

	paths := append(
		systemFontPaths,
		filepath.Join(home, "Library/Fonts/SymbolsNerdFont-Regular.ttf"),
	)

	var faces []imgfont.Face
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		f, err := truetype.Parse(data)
		if err != nil {
			continue
		}
		faces = append(faces, &indexedFace{
			Face: truetype.NewFace(f, opts),
			f:    f,
		})
	}
	return faces
}
