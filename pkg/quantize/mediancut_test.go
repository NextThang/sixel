package quantize

import (
	"image"
	"image/color"
	"testing"
)

func TestQuantizer_Quantize(t *testing.T) {
	t.Run("returns a palette with the correct number of colors", func(t *testing.T) {
		width, height := 16, 16
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := range height {
			for x := range width {
				c := color.RGBA{uint8(x * 16), uint8(y * 16), 0, 255}
				img.Set(x, y, c)
			}
		}

		numColors := 16
		quantizer := quantizer{}
		palette := make(color.Palette, 0, numColors)
		resultPalette := quantizer.Quantize(palette, img)

		if len(resultPalette) > numColors {
			t.Errorf("Expected palette to have at most %d colors, but got %d", numColors, len(resultPalette))
		}

		if len(resultPalette) == 0 {
			t.Errorf("Expected palette to have more than 0 colors, but got 0")
		}
	})

	t.Run("returns a palette with fewer colors than requested if the image has fewer unique colors", func(t *testing.T) {
		imgWithFewerColors := image.NewRGBA(image.Rect(0, 0, 2, 2))
		imgWithFewerColors.Set(0, 0, color.RGBA{255, 0, 0, 255})
		imgWithFewerColors.Set(0, 1, color.RGBA{0, 255, 0, 255})
		imgWithFewerColors.Set(1, 0, color.RGBA{0, 0, 255, 255})
		imgWithFewerColors.Set(1, 1, color.RGBA{255, 255, 0, 255})

		numColors := 8
		quantizer := quantizer{}
		palette := make(color.Palette, 0, numColors)
		resultPalette := quantizer.Quantize(palette, imgWithFewerColors)

		if len(resultPalette) != 4 {
			t.Errorf("Expected palette to have 4 colors, but got %d", len(resultPalette))
		}
	})
}

func TestQuantizer_Quantize_PanicWithPredefinedPalette(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	quantizer := quantizer{}
	palette := make(color.Palette, 1)
	palette[0] = color.Black
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	quantizer.Quantize(palette, img)
}
