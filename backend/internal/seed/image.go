package seed

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

const (
	SeedImageWidth  = 640
	SeedImageHeight = 360
)

var imagePalettes = [][3]color.RGBA{
	{{26, 35, 126, 255}, {66, 165, 245, 255}, {179, 229, 252, 255}},
	{{74, 20, 140, 255}, {171, 71, 188, 255}, {225, 190, 231, 255}},
	{{0, 77, 64, 255}, {38, 166, 154, 255}, {178, 223, 219, 255}},
	{{230, 81, 0, 255}, {255, 167, 38, 255}, {255, 224, 178, 255}},
	{{136, 14, 79, 255}, {236, 64, 122, 255}, {248, 187, 208, 255}},
	{{51, 71, 86, 255}, {120, 144, 156, 255}, {207, 216, 220, 255}},
	{{27, 94, 32, 255}, {102, 187, 106, 255}, {200, 230, 201, 255}},
	{{191, 54, 12, 255}, {255, 112, 67, 255}, {255, 204, 188, 255}},
}

// RenderImage creates compact deterministic artwork for a seeded post. The
// output is a real PNG, so seeded media exercises the same delivery path as an
// uploaded user image without relying on a third-party image service.
func RenderImage(postOrdinal, position, topic int, randomSeed int64) ([]byte, error) {
	palette := imagePalettes[topic%len(imagePalettes)]
	canvas := image.NewRGBA(image.Rect(0, 0, SeedImageWidth, SeedImageHeight))
	for y := 0; y < SeedImageHeight; y++ {
		ratio := float64(y) / float64(SeedImageHeight-1)
		shade := blend(palette[0], palette[1], ratio)
		draw.Draw(canvas, image.Rect(0, y, SeedImageWidth, y+1), &image.Uniform{C: shade}, image.Point{}, draw.Src)
	}

	state := uint64(randomSeed) ^ uint64(postOrdinal+1)*0x9e3779b97f4a7c15 ^ uint64(position+1)*0xbf58476d1ce4e5b9
	for i := 0; i < 10; i++ {
		state = state*6364136223846793005 + 1442695040888963407
		x := int(state % SeedImageWidth)
		state = state*6364136223846793005 + 1442695040888963407
		y := int(state % SeedImageHeight)
		state = state*6364136223846793005 + 1442695040888963407
		size := 24 + int(state%120)
		accent := palette[2]
		accent.A = uint8(70 + i*12)
		draw.Draw(canvas, image.Rect(x-size, y-size, x+size, y+size), &image.Uniform{C: accent}, image.Point{}, draw.Over)
	}

	// A small visual signature makes gallery images visibly distinct.
	barWidth := 50 + (postOrdinal*17+position*31)%220
	draw.Draw(canvas, image.Rect(32, SeedImageHeight-54, 32+barWidth, SeedImageHeight-30), &image.Uniform{C: color.RGBA{255, 255, 255, 210}}, image.Point{}, draw.Over)
	draw.Draw(canvas, image.Rect(32, SeedImageHeight-24, 132+position*30, SeedImageHeight-16), &image.Uniform{C: palette[2]}, image.Point{}, draw.Over)

	var output bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(&output, canvas); err != nil {
		return nil, fmt.Errorf("encode seed image: %w", err)
	}
	return output.Bytes(), nil
}

func blend(a, b color.RGBA, ratio float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-ratio) + float64(b.R)*ratio),
		G: uint8(float64(a.G)*(1-ratio) + float64(b.G)*ratio),
		B: uint8(float64(a.B)*(1-ratio) + float64(b.B)*ratio),
		A: 255,
	}
}
