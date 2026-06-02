package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	f, err := os.Create("tmp/generated.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
