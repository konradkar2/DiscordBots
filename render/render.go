package render

import (
	"fmt"
	"image"
	"image/color"
	imagedraw "image/draw"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type RenderOptions struct {
	PaddingX       int
	PaddingY       int
	OverlayAlpha   uint8
	TextSize       float64
	MinTextSize    float64
	AuthorSize     float64
	MinAuthorSize  float64
	LineSpacing    int
	AuthorSpacing  int
	TextColor      color.Color
	AuthorColor    color.Color
	BackgroundPath string
	FontPath       string
	OutputPath     string
}

func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		PaddingX:      80,
		PaddingY:      60,
		OverlayAlpha:  110,
		TextSize:      42,
		MinTextSize:   18,
		AuthorSize:    24,
		MinAuthorSize: 14,
		LineSpacing:   14,
		AuthorSpacing: 28,
		TextColor:     color.RGBA{245, 245, 240, 255},
		AuthorColor:   color.RGBA{220, 220, 215, 255},
	}
}

func RenderQuoteCard(text, author string, opts RenderOptions) error {
	bg, err := loadImage(opts.BackgroundPath)
	if err != nil {
		return fmt.Errorf("load background: %w", err)
	}

	bounds := bg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	imagedraw.Draw(canvas, canvas.Bounds(), bg, bounds.Min, imagedraw.Src)

	applyOverlay(canvas, color.RGBA{0, 0, 0, opts.OverlayAlpha})

	fontBytes, err := os.ReadFile(opts.FontPath)
	if err != nil {
		return fmt.Errorf("read font: %w", err)
	}

	tt, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("parse font: %w", err)
	}

	textFace, finalTextSize, lines, err := fitTextToBox(
		tt,
		text,
		opts.TextSize,
		opts.MinTextSize,
		width-2*opts.PaddingX,
		height-2*opts.PaddingY,
		opts.LineSpacing,
		author != "",
		opts.AuthorSize,
		opts.MinAuthorSize,
		opts.AuthorSpacing,
	)
	if err != nil {
		return err
	}
	defer closeFace(textFace)

	var authorFace font.Face
	if strings.TrimSpace(author) != "" {
		authorFace, err = fitSingleLineFace(
			tt,
			"— "+author,
			opts.AuthorSize,
			opts.MinAuthorSize,
			width-2*opts.PaddingX,
		)
		if err != nil {
			return err
		}
		defer closeFace(authorFace)
	}

	textBlockHeight := blockHeight(textFace, len(lines), opts.LineSpacing)

	authorLine := ""
	authorHeight := 0
	if strings.TrimSpace(author) != "" {
		authorLine = "— " + author
		m := authorFace.Metrics()
		authorHeight = (m.Ascent + m.Descent).Ceil()
	}

	totalHeight := textBlockHeight
	if authorLine != "" {
		totalHeight += opts.AuthorSpacing + authorHeight
	}

	startY := (height-totalHeight)/2 + textFace.Metrics().Ascent.Ceil()

	drawCenteredLines(canvas, lines, textFace, opts.TextColor, width/2, startY, opts.LineSpacing)

	if authorLine != "" {
		authorY := startY + textBlockHeight + opts.AuthorSpacing
		drawCenteredLine(canvas, authorLine, authorFace, opts.AuthorColor, width/2, authorY)
	}

	out, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, canvas); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	_ = finalTextSize
	return nil
}

func fitTextToBox(
	tt *opentype.Font,
	text string,
	startSize float64,
	minSize float64,
	maxWidth int,
	maxHeight int,
	lineSpacing int,
	hasAuthor bool,
	authorStartSize float64,
	authorMinSize float64,
	authorSpacing int,
) (font.Face, float64, []string, error) {
	for size := startSize; size >= minSize; size -= 1 {
		textFace, err := newFace(tt, size)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("create text face: %w", err)
		}

		lines := wrapText(text, textFace, maxWidth)
		if len(lines) == 0 {
			lines = []string{""}
		}

		tooWide := false
		for _, line := range lines {
			if textWidth(textFace, line) > maxWidth {
				tooWide = true
				break
			}
		}
		if tooWide {
			closeFace(textFace)
			continue
		}

		totalHeight := blockHeight(textFace, len(lines), lineSpacing)

		if hasAuthor {
			authorFace, err := fitSingleLineFace(tt, "— author", authorStartSize, authorMinSize, maxWidth)
			if err != nil {
				closeFace(textFace)
				return nil, 0, nil, err
			}
			am := authorFace.Metrics()
			totalHeight += authorSpacing + (am.Ascent+am.Descent).Ceil()
			closeFace(authorFace)
		}

		if totalHeight <= maxHeight {
			return textFace, size, lines, nil
		}

		closeFace(textFace)
	}

	return nil, 0, nil, fmt.Errorf("text does not fit even at minimum font size")
}

func fitSingleLineFace(
	tt *opentype.Font,
	text string,
	startSize float64,
	minSize float64,
	maxWidth int,
) (font.Face, error) {
	for size := startSize; size >= minSize; size -= 1 {
		face, err := newFace(tt, size)
		if err != nil {
			return nil, fmt.Errorf("create face: %w", err)
		}
		if textWidth(face, text) <= maxWidth {
			return face, nil
		}
		closeFace(face)
	}
	return nil, fmt.Errorf("single line text does not fit even at minimum font size")
}

func newFace(tt *opentype.Font, size float64) (font.Face, error) {
	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func applyOverlay(img *image.RGBA, c color.RGBA) {
	imagedraw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, imagedraw.Over)
}

func wrapText(text string, face font.Face, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		candidate := current + " " + word
		if textWidth(face, candidate) <= maxWidth {
			current = candidate
		} else {
			lines = append(lines, current)
			current = word
		}
	}

	lines = append(lines, current)
	return lines
}

func textWidth(face font.Face, s string) int {
	var d font.Drawer
	d.Face = face
	return d.MeasureString(s).Ceil()
}

func blockHeight(face font.Face, lineCount int, lineSpacing int) int {
	if lineCount <= 0 {
		return 0
	}
	m := face.Metrics()
	lineHeight := (m.Ascent + m.Descent).Ceil()
	return lineCount*lineHeight + (lineCount-1)*lineSpacing
}

func drawCenteredLines(img *image.RGBA, lines []string, face font.Face, col color.Color, centerX, startY, lineSpacing int) {
	m := face.Metrics()
	lineHeight := (m.Ascent + m.Descent).Ceil()

	y := startY
	for _, line := range lines {
		drawCenteredLine(img, line, face, col, centerX, y)
		y += lineHeight + lineSpacing
	}
}

func drawCenteredLine(img *image.RGBA, line string, face font.Face, col color.Color, centerX, baselineY int) {
	w := textWidth(face, line)
	x := centerX - w/2

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(x, baselineY),
	}
	d.DrawString(line)
}

func closeFace(face font.Face) {
	type closer interface {
		Close() error
	}
	if c, ok := face.(closer); ok {
		_ = c.Close()
	}
}