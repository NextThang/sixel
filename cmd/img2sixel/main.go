package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/nextthang/sixel"
)

func main() {
	reader := os.Stdin
	if len(os.Args) > 1 {
		reader, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer reader.Close()
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding image: %v\n", err)
		os.Exit(1)
	}

	if err := sixel.Encode(os.Stdout, img); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding image to sixel: %v\n", err)
		os.Exit(1)
	}
}
