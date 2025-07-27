# Go Sixel Encoder

A Go library and command-line tool for encoding images into the [Sixel](https://en.wikipedia.org/wiki/Sixel) graphics format. This allows for displaying images directly within compatible terminal emulators.

This project provides both a reusable library and a simple command-line utility, `img2sixel`, for converting images.

## Features

- **Sixel Encoding**: Converts standard image formats (PNG, JPEG, GIF) into Sixel data.
- **Color Quantization**: Automatically creates an optimized 256-color palette for any input image using a Median Cut algorithm.
- **Dithering**: Applies Floyd-Steinberg dithering to produce higher-quality output that more closely resembles the original image.
- **Command-Line Tool**: Includes the `img2sixel` utility for easy file conversion from the shell.
- **Go Library**: Exposes a simple `Encode` function for use in your own Go applications.

## Installation

### Command-Line Tool

To install the `img2sixel` command-line tool, run:

```sh
go install github.com/nextthang/sixel/cmd/img2sixel@latest
```

This will place the `img2sixel` binary in your `$GOPATH/bin` directory.

### Library

To use the Sixel encoder as a library in your own project, add it as a dependency:

```sh
go get github.com/nextthang/sixel
```

## Usage

### `img2sixel` CLI

The command-line tool can read an image from a file path or from standard input. The resulting Sixel data is written to standard output.

**From a file:**

```sh
# This will print the Sixel output directly to your terminal.
# If your terminal is Sixel-compatible, it will display the image.
img2sixel path/to/your/image.png
```

**From standard input:**

```sh
cat path/to/your/image.jpg | img2sixel
```

### Library Usage

You can use the `sixel` package in your own Go code to encode any `image.Image` object.

Here is a basic example:

```go
package main

import (
 "fmt"
 "image"
 _ "image/png" // Register the PNG decoder
 "os"

 "github.com/nextthang/sixel"
)

func main() {
 // Open an image file
 file, err := os.Open("path/to/your/image.png")
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error opening file: %v
", err)
  os.Exit(1)
 }
 defer file.Close()

 // Decode the image
 img, _, err := image.Decode(file)
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error decoding image: %v
", err)
  os.Exit(1)
 }

 // Encode the image to Sixel format and write to standard output
 if err := sixel.Encode(os.Stdout, img); err != nil {
  fmt.Fprintf(os.Stderr, "Error encoding image to sixel: %v
", err)
  os.Exit(1)
 }
}
```

### Using with `image/gif`

The median cut implementation in the `quantize` package implements the `draw.Quantizer` interface. This means you can use it with other packages in the standard library that accept a `draw.Quantizer`, such as the `image/gif` package's encoder.

Here is an example of how to use the `quantize.Quantizer` when encoding a GIF:

```go
package main

import (
 "fmt"
 "image"
 "image/gif"
 _ "image/png"
 "os"

 "github.com/nextthang/sixel/quantize"
)

func main() {
 // Open an image file
 file, err := os.Open("path/to/your/image.png")
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error opening file: %v
", err)
  os.Exit(1)
 }
 defer file.Close()

 // Decode the image
 img, _, err := image.Decode(file)
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error decoding image: %v
", err)
  os.Exit(1)
 }

 // Create a new file to write the GIF to
 outFile, err := os.Create("output.gif")
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error creating file: %v
", err)
  os.Exit(1)
 }
 defer outFile.Close()

 // Encode the image to a GIF using the custom quantizer
 err = gif.Encode(outFile, img, &gif.Options{
  NumColors: 256,
  Quantizer: quantize.Quantizer,
 })
 if err != nil {
  fmt.Fprintf(os.Stderr, "Error encoding GIF: %v
", err)
  os.Exit(1)
 }
}
```

## License

This project is licensed under the terms of the `LICENSE` file.
