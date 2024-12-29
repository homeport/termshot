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
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/term"
	"github.com/homeport/termshot/internal/fonts"
	"golang.org/x/image/font"
)

const (
	red    = "#ED655A"
	yellow = "#E1C04C"
	green  = "#71BD47"
)

type Scaffold struct {
	content bunt.String

	factor float64

	columns int

	defaultForegroundColor color.Color

	drawDecorations bool
	drawShadow      bool

	shadowBaseColor string
	shadowRadius    uint8
	shadowOffsetX   float64
	shadowOffsetY   float64

	padding float64
	margin  float64

	regular     font.Face
	bold        font.Face
	italic      font.Face
	boldItalic  font.Face
	lineSpacing float64
	tabSpaces   int
}

func NewImageCreator() Scaffold {
	f := 2.0

	fontRegular, _ := truetype.Parse(fonts.HackRegular)
	fontBold, _ := truetype.Parse(fonts.HackBold)
	fontItalic, _ := truetype.Parse(fonts.HackItalic)
	fontBoldItalic, _ := truetype.Parse(fonts.HackBoldItalic)
	fontFaceOptions := &truetype.Options{Size: f * 12, DPI: 144}

	return Scaffold{
		defaultForegroundColor: bunt.LightGray,

		factor: f,

		margin:  f * 48,
		padding: f * 24,

		drawDecorations: true,
		drawShadow:      true,

		shadowBaseColor: "#10101066",
		shadowRadius:    uint8(math.Min(f*16, 255)),
		shadowOffsetX:   f * 16,
		shadowOffsetY:   f * 16,

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

func (s *Scaffold) SetColumns(columns int) { s.columns = columns }

func (s *Scaffold) DrawDecorations(value bool) { s.drawDecorations = value }

func (s *Scaffold) DrawShadow(value bool) { s.drawShadow = value }

func (s *Scaffold) GetFixedColumns() int {
	if s.columns != 0 {
		return s.columns
	}

	columns, _ := term.GetTerminalSize()
	return columns
}

func (s *Scaffold) AddContent(in io.Reader) error {
	parsed, err := bunt.ParseStream(in)
	if err != nil {
		return fmt.Errorf("failed to parse input stream: %w", err)
	}

	var tmp bunt.String
	var counter int
	for _, cr := range *parsed {
		counter++

		if cr.Symbol == '\n' {
			counter = 0
		}

		// Add an additional newline in case the column
		// count is reached and line wrapping is needed
		if counter > s.GetFixedColumns() {
			counter = 0
			tmp = append(tmp, bunt.ColoredRune{
				Settings: cr.Settings,
				Symbol:   '\n',
			})
		}

		tmp = append(tmp, cr)
	}

	s.content = append(s.content, tmp...)

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

	// temporary drawer for reference calucation
	tmpDrawer := &font.Drawer{Face: s.regular}

	// width, either by using longest line, or by fixed column value
	switch s.columns {
	case 0: // unlimited: max width of all lines
		for _, line := range lines {
			advance := tmpDrawer.MeasureString(line)
			if lineWidth := float64(advance >> 6); lineWidth > width {
				width = lineWidth
			}
		}

	default: // fixed: max width based on column count
		width = float64(tmpDrawer.MeasureString(strings.Repeat("a", s.GetFixedColumns())) >> 6)
	}

	// height, lines times font height and line spacing
	height = float64(len(lines)) * s.fontHeight() * s.lineSpacing

	return width, height
}

func (s *Scaffold) image() (image.Image, error) {
	var f = func(value float64) float64 { return s.factor * value }

	var (
		corner   = f(6)
		radius   = f(9)
		distance = f(25)
	)

	contentWidth, contentHeight := s.measureContent()

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered
	contentWidth = math.Max(contentWidth, 3*distance+3*radius)

	marginX, marginY := s.margin, s.margin
	paddingX, paddingY := s.padding, s.padding

	xOffset := marginX
	yOffset := marginY

	var titleOffset float64
	if s.drawDecorations {
		titleOffset = f(40)
	}

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	dc := gg.NewContext(int(width), int(height))

	// Optional: Apply blurred rounded rectangle to mimic the window shadow
	//
	if s.drawShadow {
		xOffset -= s.shadowOffsetX / 2
		yOffset -= s.shadowOffsetY / 2

		bc := gg.NewContext(int(width), int(height))
		bc.DrawRoundedRectangle(xOffset+s.shadowOffsetX, yOffset+s.shadowOffsetY, width-2*marginX, height-2*marginY, corner)
		bc.SetHexColor(s.shadowBaseColor)
		bc.Fill()

		shadow, err := stackblur.Process(bc.Image(), uint32(s.shadowRadius))
		if err != nil {
			return nil, err
		}

		dc.DrawImage(shadow, 0, 0)
	}

	// Draw rounded rectangle with outline to produce impression of a window
	//
	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor("#151515")
	dc.Fill()

	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor("#404040")
	dc.SetLineWidth(f(1))
	dc.Stroke()

	// Optional: Draw window decorations (i.e. three buttons) to produce the
	// impression of an actional window
	//
	if s.drawDecorations {
		for i, color := range []string{red, yellow, green} {
			dc.DrawCircle(xOffset+paddingX+float64(i)*distance+f(4), yOffset+paddingY+f(4), radius)
			dc.SetHexColor(color)
			dc.Fill()
		}
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

		str := string(cr.Symbol)
		w, h := dc.MeasureString(str)

		// background color
		switch cr.Settings & 0x02 {
		case 2:
			dc.SetRGB255(
				int((cr.Settings>>32)&0xFF),
				int((cr.Settings>>40)&0xFF),
				int((cr.Settings>>48)&0xFF),
			)

			dc.DrawRectangle(x, y-h+12, w, h)
			dc.Fill()
		}

		// foreground color
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

		switch str {
		case "\n":
			x = xOffset + paddingX
			y += h * s.lineSpacing
			continue

		case "\t":
			x += w * float64(s.tabSpaces)
			continue

		case "✗", "ˣ": // mitigate issue #1 by replacing it with a similar character
			str = "×"
		}

		dc.DrawString(str, x, y)

		// There seems to be no font face based way to do an underlined
		// string, therefore manually draw a line under each character
		if cr.Settings&0x1C == 16 {
			dc.DrawLine(x, y+f(4), x+w, y+f(4))
			dc.SetLineWidth(f(1))
			dc.Stroke()
		}

		x += w
	}

	return dc.Image(), nil
}

func (s *Scaffold) Write(w io.Writer) error {
	image, err := s.image()
	if err != nil {
		return err
	}

	return png.Encode(w, image)
}
