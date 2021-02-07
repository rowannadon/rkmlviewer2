package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
)

func loadImage(filepath string) ([]uint8, int32, int32) {
	// You can register another format here
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)

	file, err := os.Open(filepath)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	img, _, err := image.Decode(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	// b := img.Bounds()

	// var pixels []float32
	// for y := b.Min.Y; y < b.Max.Y; y++ {
	// 	for x := b.Min.X; x < b.Max.X; x++ {
	// 		r, g, b, a := img.At(x, y).RGBA()

	// 		pixels = append(pixels, float32(r)/float32(a), float32(g)/float32(a), float32(b)/float32(a))
	// 	}
	// }

	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, rect.Min, draw.Src)

	return rgba.Pix, int32(rect.Max.X - rect.Min.X), int32(rect.Max.Y - rect.Min.Y)
}
