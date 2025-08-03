package sixel

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"strings"

	"github.com/nextthang/sixel/pkg/quantize"
)

const (
	header      = "\x1bP" // There's also a DCS variant (\x90) but it's less commonly supported
	endOfHeader = 'q'
	paletteSize = 256 // Some terminals support more, but 256 is a common limit
	bandHeight  = 6
	sixelOffset = 63
	terminator  = "\x1b\\"
)

type (
	void       struct{}
	paletteMap map[color.Color]uint8
	colorSet   map[color.Color]void
)

func (pm paletteMap) String() string {
	var result strings.Builder
	for c, idx := range pm {
		r, g, b, _ := c.RGBA()

		rScale := ((r >> 8) * 100) / 255
		gScale := ((g >> 8) * 100) / 255
		bScale := ((b >> 8) * 100) / 255

		fmt.Fprintf(&result, "#%d;2;%d;%d;%d", idx, rScale, gScale, bScale)
	}
	return result.String()
}

func createPaletteMap(palette color.Palette) paletteMap {
	if len(palette) > paletteSize {
		panic(fmt.Sprintf("Palette size exceeds the maximum limit of %d colors", paletteSize))
	}

	pm := make(paletteMap, len(palette))
	for idx, c := range palette {
		pm[c] = uint8(idx)
	}
	return pm
}

func rleGenerateSubString(char rune, count int) string {
	if count > 2 {
		return fmt.Sprintf("!%d%c", count, char)
	}
	return strings.Repeat(string(char), count)
}

func rleEncode(input string) string {
	if len(input) == 0 {
		return ""
	}

	var result strings.Builder
	count := 1
	for i := 1; i < len(input); i++ {
		if input[i] == input[i-1] {
			count++
		} else {
			result.WriteString(rleGenerateSubString(rune(input[i-1]), count))
			count = 1
		}
	}
	result.WriteString(rleGenerateSubString(rune(input[len(input)-1]), count))

	return result.String()
}

func findUniqueColorsInBand(img image.Image, bounds image.Rectangle, yStart int) colorSet {
	uniqueColors := make(colorSet)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := yStart; y < yStart+bandHeight && y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			uniqueColors[c] = void{}
		}
	}
	return uniqueColors
}

func Encode(w io.Writer, img image.Image) error {
	bounds := img.Bounds()
	palettedImage, ok := img.(*image.Paletted)
	if palettedImage == nil || !ok || len(palettedImage.Palette) > paletteSize {
		palettedImage = image.NewPaletted(bounds, make(color.Palette, 0, paletteSize))
		palettedImage.Palette = quantize.Quantizer.Quantize(palettedImage.Palette, img)
		draw.FloydSteinberg.Draw(palettedImage, bounds, img, bounds.Min)
	}

	paletteMap := createPaletteMap(palettedImage.Palette)

	writer := bufio.NewWriter(w)
	defer writer.Flush()

	writer.WriteString(header)
	writer.WriteRune(endOfHeader)
	// Setting the raster attributes to 1:1 pixel aspect ratio and width/height of the image.
	fmt.Fprintf(writer, "\"1;1;%d;%d", bounds.Dx(), bounds.Dy())
	writer.WriteString(paletteMap.String())

	for curY := bounds.Min.Y; curY < bounds.Max.Y; curY += bandHeight {
		uniqueColors := findUniqueColorsInBand(palettedImage, bounds, curY)
		index := 0
		for c := range uniqueColors {
			idx := paletteMap[c]
			if _, err := fmt.Fprintf(writer, "#%d", idx); err != nil {
				return err
			}

			var line strings.Builder
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				var pixel int
				for i := range bandHeight {
					y := curY + i
					if y >= bounds.Max.Y {
						break
					}
					if c == palettedImage.At(x, y) {
						pixel |= (1 << i)
					}
				}
				pixel += sixelOffset
				line.WriteRune(rune(pixel))
			}
			writer.WriteString(rleEncode(line.String()))
			index++
			if index < len(uniqueColors) {
				writer.WriteRune('$') // Carriage Return
			}
		}
		if curY+bandHeight < bounds.Max.Y {
			writer.WriteRune('-') // New Line
		}
	}
	writer.WriteString(terminator)

	return nil
}
