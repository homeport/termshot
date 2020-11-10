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
	"image/color"
	"math"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/term"
	"golang.org/x/image/font"
)

const (
	red    = "#ED655A"
	yellow = "#E1C04C"
	green  = "#71BD47"
)

type Scaffold struct {
	content bunt.String

	columns int
	rows    int

	defaultForegroundColor color.Color

	drawShadow      bool
	shadowBaseColor string
	shadowRadius    uint8
	shadowOffsetX   int
	shadowOffsetY   int

	padding int
	margin  int

	regular     font.Face
	bold        font.Face
	italic      font.Face
	boldItalic  font.Face
	lineSpacing float64
	tabSpaces   int
}

func NewImageCreator() Scaffold {
	var loadFont = func(assetName string) *truetype.Font {
		bytes, _ := Asset(assetName)
		font, _ := truetype.Parse(bytes)
		return font
	}

	cols, rows := term.GetTerminalSize()

	fontRegular := loadFont("Hack-Regular.ttf")
	fontBold := loadFont("Hack-Bold.ttf")
	fontItalic := loadFont("Hack-Italic.ttf")
	fontBoldItalic := loadFont("Hack-BoldItalic.ttf")
	fontFaceOptions := &truetype.Options{Size: 48, DPI: 144}

	return Scaffold{
		defaultForegroundColor: bunt.LightGray,

		columns: cols,
		rows:    rows,

		margin:  196,
		padding: 96,

		drawShadow:      true,
		shadowBaseColor: "#10101066",
		shadowRadius:    64,
		shadowOffsetX:   64,
		shadowOffsetY:   64,

		regular:     truetype.NewFace(fontRegular, fontFaceOptions),
		bold:        truetype.NewFace(fontBold, fontFaceOptions),
		italic:      truetype.NewFace(fontItalic, fontFaceOptions),
		boldItalic:  truetype.NewFace(fontBoldItalic, fontFaceOptions),
		lineSpacing: 1.2,
		tabSpaces:   2,
	}
}

func (s *Scaffold) SetFontFaceRegular(face font.Face) { s.regular = face }

func (s *Scaffold) SetFontFaceBold(face font.Face) { s.bold = face }

func (s *Scaffold) SetFontFaceItalic(face font.Face) { s.italic = face }

func (s *Scaffold) SetFontFaceBoldItalic(face font.Face) { s.boldItalic = face }

func (s *Scaffold) AddContent(content string) error {
	tmp, err := bunt.ParseString(content)
	if err != nil {
		return err
	}

	s.content = append(s.content, *tmp...)

	return nil
}

func (s *Scaffold) fontHeight() float64 {
	return float64(s.regular.Metrics().Height >> 6)
}

func (s *Scaffold) measureContent() (width float64, height float64) {
	var tmp = make([]rune, len(s.content))
	for i, cr := range s.content {
		tmp[i] = cr.Symbol
	}

	lines := strings.Split(
		strings.TrimSuffix(
			string(tmp),
			"\n",
		),
		"\n",
	)

	// width, max width of all lines
	tmpDrawer := &font.Drawer{Face: s.regular}
	for _, line := range lines {
		advance := tmpDrawer.MeasureString(line)
		if lineWidth := float64(advance >> 6); lineWidth > width {
			width = lineWidth
		}
	}

	// height, lines times font height and line spacing
	height = float64(len(lines)) * s.fontHeight() * s.lineSpacing

	return width, height
}

func (s *Scaffold) SavePNG(path string) error {
	const (
		corner   = 24
		radius   = 36
		distance = 100
	)

	contentWidth, contentHeight := s.measureContent()

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered
	contentWidth = math.Max(contentWidth, 3*distance+3*radius)

	marginX, marginY := float64(s.margin), float64(s.margin)
	paddingX, paddingY := float64(s.padding), float64(s.padding)

	xOffset := marginX
	yOffset := marginY
	titleOffset := 160.0

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	dc := gg.NewContext(int(width), int(height))

	// Optional: Apply blurred rounded rectangle to mimic the window shadow
	//
	if s.drawShadow {
		xOffset -= float64(s.shadowOffsetX >> 1)
		yOffset -= float64(s.shadowOffsetY >> 1)

		bc := gg.NewContext(int(width), int(height))

		bc.DrawRoundedRectangle(xOffset+float64(s.shadowOffsetX), yOffset+float64(s.shadowOffsetY), width-2*marginX, height-2*marginY, corner)
		bc.SetHexColor(s.shadowBaseColor)
		bc.Fill()

		var done = make(chan struct{}, s.shadowRadius)
		shadow := stackblur.Process(
			bc.Image(),
			uint32(width),
			uint32(height),
			uint32(s.shadowRadius),
			done,
		)

		<-done
		dc.DrawImage(shadow, 0, 0)
	}

	// Draw rounded rectangle with outline and three button to produce the
	// impression of a window with controls and a content area
	//
	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor("#151515")
	dc.Fill()

	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor("#404040")
	dc.SetLineWidth(4)
	dc.Stroke()

	for i, color := range []string{red, yellow, green} {
		dc.DrawCircle(xOffset+paddingX+float64(i*distance)+16, yOffset+paddingY+16, radius)
		dc.SetHexColor(color)
		dc.Fill()
	}

	// Apply the actual text into the prepared content area of the window
	//
	var x, y float64 = xOffset + paddingX, yOffset + paddingY + titleOffset + s.fontHeight()
	for _, cr := range s.content {
		switch cr.Settings & 0x1C {
		case 4:
			dc.SetFontFace(s.bold)

		case 8:
			dc.SetFontFace(s.italic)

		case 12:
			dc.SetFontFace(s.boldItalic)

		default:
			dc.SetFontFace(s.regular)
		}

		switch cr.Settings & 0x01 {
		case 1:
			dc.SetRGB255(
				int((cr.Settings>>8)&0xFF),
				int((cr.Settings>>16)&0xFF),
				int((cr.Settings>>24)&0xFF),
			)

		default:
			dc.SetColor(s.defaultForegroundColor)
		}

		str := string(cr.Symbol)
		w, h := dc.MeasureString(str)

		switch str {
		case "\n":
			x = xOffset + paddingX
			y += h * s.lineSpacing
			continue

		case "\t":
			x += w * float64(s.tabSpaces)
			continue
		}

		dc.DrawString(str, x, y)

		// There seems to be no font face based way to do an underlined
		// string, therefore manually draw a line under each character
		if cr.Settings&0x1C == 16 {
			dc.DrawLine(x, y+16, x+w, y+16)
			dc.SetLineWidth(4)
			dc.Stroke()
		}

		x += w
	}

	return dc.SavePNG(path)
}
