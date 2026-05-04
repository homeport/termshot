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
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// glyphPresence is implemented by faces backed by a *truetype.Font.
// It gives an authoritative answer about whether a rune is in the font's
// cmap, which the font.Face ok returns cannot — truetype faces return ok=true
// even for missing runes because they succeed in rasterizing glyph 0 (.notdef).
type glyphPresence interface {
	hasGlyph(r rune) bool
}

// FallbackFace wraps a primary font.Face with a slice of fallback faces.
//
// If the primary implements glyphPresence, it is checked first; glyphs
// present in the primary are rendered by the primary, and only missing glyphs
// fall through to the fallback list (then to primary's .notdef tofu).
//
// If the primary does not implement glyphPresence, it is used directly for
// all glyphs (no fallback is consulted), preserving pre-fallback behavior.
type FallbackFace struct {
	primary   font.Face
	fallbacks []font.Face
}

func NewFallbackFace(primary font.Face, fallbacks ...font.Face) *FallbackFace {
	return &FallbackFace{primary: primary, fallbacks: fallbacks}
}

func (f *FallbackFace) Close() error {
	for _, fb := range f.fallbacks {
		fb.Close()
	}
	return f.primary.Close()
}

func (f *FallbackFace) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	if gp, isGP := f.primary.(glyphPresence); isGP {
		if gp.hasGlyph(r) {
			return f.primary.Glyph(dot, r)
		}
		for _, fb := range f.fallbacks {
			if gp2, isGP2 := fb.(glyphPresence); isGP2 && !gp2.hasGlyph(r) {
				continue
			}
			if dr, mask, maskp, advance, ok = fb.Glyph(dot, r); ok {
				return
			}
		}
		dr, mask, maskp, advance, ok = f.primary.Glyph(dot, r)
		return
	}
	dr, mask, maskp, advance, ok = f.primary.Glyph(dot, r)
	return
}

func (f *FallbackFace) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	if gp, isGP := f.primary.(glyphPresence); isGP {
		if gp.hasGlyph(r) {
			return f.primary.GlyphBounds(r)
		}
		for _, fb := range f.fallbacks {
			if gp2, isGP2 := fb.(glyphPresence); isGP2 && !gp2.hasGlyph(r) {
				continue
			}
			if bounds, advance, ok = fb.GlyphBounds(r); ok {
				return
			}
		}
		bounds, advance, ok = f.primary.GlyphBounds(r)
		return
	}
	bounds, advance, ok = f.primary.GlyphBounds(r)
	return
}

func (f *FallbackFace) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	if gp, isGP := f.primary.(glyphPresence); isGP {
		if gp.hasGlyph(r) {
			return f.primary.GlyphAdvance(r)
		}
		for _, fb := range f.fallbacks {
			if gp2, isGP2 := fb.(glyphPresence); isGP2 && !gp2.hasGlyph(r) {
				continue
			}
			if advance, ok = fb.GlyphAdvance(r); ok {
				return
			}
		}
		advance, ok = f.primary.GlyphAdvance(r)
		return
	}
	advance, ok = f.primary.GlyphAdvance(r)
	return
}

func (f *FallbackFace) Kern(r0, r1 rune) fixed.Int26_6 {
	return f.primary.Kern(r0, r1)
}

func (f *FallbackFace) Metrics() font.Metrics {
	return f.primary.Metrics()
}
