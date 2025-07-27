package quantize

import (
	"image"
	"image/color"
	"slices"
)

const (
	rIdx = iota
	gIdx
	bIdx
)

const supportedDimensions = 3

type (
	void     struct{}
	colorSet map[rgbColor]void
	bucket   []rgbColor
	rgbColor [supportedDimensions]uint8
)

type Quantizer struct{}

func (cs colorSet) fillColors(img image.Image) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			cs[rgbColor{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}] = void{}
		}
	}
}

func (cs colorSet) toBucket() bucket {
	b := make(bucket, 0, len(cs))
	for c := range cs {
		b = append(b, c)
	}
	return b
}

func (cs colorSet) toColorSlice() []color.Color {
	slice := make([]color.Color, 0, len(cs))
	for c := range cs {
		slice = append(slice, color.RGBA{c[rIdx], c[gIdx], c[bIdx], 255})
	}
	return slice
}

func (b bucket) longestRangeDimension() int {
	maxVals := make([]uint8, supportedDimensions)
	minVals := make([]uint8, supportedDimensions)

	for dim := range supportedDimensions {
		minVals[dim] = 255
		maxVals[dim] = 0
	}

	for _, curCol := range b {
		for dim := range supportedDimensions {
			if curCol[dim] < minVals[dim] {
				minVals[dim] = curCol[dim]
			}
			if curCol[dim] > maxVals[dim] {
				maxVals[dim] = curCol[dim]
			}
		}
	}

	longestRangeDim := 0
	longestRange := maxVals[0] - minVals[0]
	for dim := 1; dim < supportedDimensions; dim++ {
		r := maxVals[dim] - minVals[dim]
		if r > longestRange {
			longestRangeDim = dim
			longestRange = r
		}
	}

	return longestRangeDim
}

func (b bucket) sortInDimension(dimension int) {
	slices.SortFunc(b, func(a, b rgbColor) int {
		return int(a[dimension]) - int(b[dimension])
	})
}

func (b bucket) split() (bucket, bucket) {
	if len(b) < 2 {
		return b, nil
	}

	b.sortInDimension(b.longestRangeDimension())
	mid := len(b) / 2

	return b[:mid], b[mid:]
}

func (c rgbColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c[rIdx])
	r |= r << 8
	g = uint32(c[gIdx])
	g |= g << 8
	b = uint32(c[bIdx])
	b |= b << 8
	a = uint32(0xffff) // No alpha channel in this quantizer
	return
}

func findLargestBucket(buckets []bucket) int {
	if len(buckets) == 0 {
		return -1
	}

	largestBucketIndex := 0
	largestBucketSize := len(buckets[0])

	for i := 1; i < len(buckets); i++ {
		if len(buckets[i]) > largestBucketSize {
			largestBucketIndex = i
			largestBucketSize = len(buckets[i])
		}
	}
	return largestBucketIndex
}

func (Quantizer) Quantize(p color.Palette, img image.Image) color.Palette {
	if len(p) > 0 {
		panic("Quantizer does not support pre-defined palettes")
	}
	numberOfColors := cap(p)
	cs := make(colorSet, numberOfColors)
	cs.fillColors(img)

	if len(cs) <= numberOfColors {
		colors := cs.toColorSlice()
		return color.Palette(colors)
	}

	buckets := make([]bucket, 0, numberOfColors)
	buckets = append(buckets, cs.toBucket())

	for len(buckets) < numberOfColors {
		largestBucketIndex := findLargestBucket(buckets)
		if largestBucketIndex < 0 {
			panic("No buckets found, cannot split further")
		}

		currentBucket := buckets[largestBucketIndex]
		newBucket1, newBucket2 := currentBucket.split()
		if newBucket2 == nil {
			panic("Bucket split resulted in nil bucket, cannot continue")
		}
		buckets[largestBucketIndex] = newBucket1
		buckets = append(buckets, newBucket2)
	}

	palette := make(color.Palette, 0, len(buckets))
	for _, bkt := range buckets {
		if len(bkt) == 0 {
			panic("Empty bucket encountered, cannot generate color")
		}

		r, g, b := uint32(0), uint32(0), uint32(0)
		for _, c := range bkt {
			r += uint32(c[rIdx])
			g += uint32(c[gIdx])
			b += uint32(c[bIdx])
		}

		n := uint32(len(bkt))
		palette = append(palette, rgbColor{uint8(r / n), uint8(g / n), uint8(b / n)})
	}

	if len(palette) > numberOfColors {
		panic("Somehow more colors than requested were generated")
	}

	return palette
}
