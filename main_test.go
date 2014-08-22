package main

import (
	"fmt"
	"image"
	"image/color"
	"testing"
)

type ArrayImage struct {
	Colors [][]color.Color
}

func (i ArrayImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{len(i.Colors), len(i.Colors[0])},
	}
}

func (i ArrayImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (i ArrayImage) At(x, y int) color.Color {
	return i.Colors[x][y]
}

func makeColors(is [][]uint8) [][]color.Color {
	arr := [][]color.Color{
		[]color.Color{nil, nil, nil},
		[]color.Color{nil, nil, nil},
		[]color.Color{nil, nil, nil},
	}
	for y := 0; y < len(is); y++ {
		for x := 0; x < len(is); x++ {
			fmt.Println(x, y)
			arr[x][y] = color.RGBA{is[x][y], 0, 0, 0}
		}
	}
	return arr
}

// unexported until I can think about this a little better.
func testRotate(t *testing.T) {

	var start = [][]uint8{
		[]uint8{1, 2, 3},
		[]uint8{4, 5, 6},
		[]uint8{7, 8, 9},
	}

	var exp = [][]uint8{
		[]uint8{7, 4, 1},
		[]uint8{8, 5, 2},
		[]uint8{9, 6, 3},
	}

	var foozie = ArrayImage{
		Colors: makeColors(start),
	}

	foozieBounds := foozie.Bounds()
	result := image.NewRGBA(foozieBounds)
	rotate(foozie, *result)

	expectedResult := ArrayImage{
		Colors: makeColors(exp),
	}

	fmt.Println(expectedResult.Colors)

	for x := foozieBounds.Min.X; x < foozieBounds.Max.X; x++ {
		for y := foozieBounds.Min.Y; y < foozieBounds.Max.Y; y++ {
			if result.At(x, y) != expectedResult.At(x, y) {
				fmt.Println(result)
				fmt.Println(expectedResult)
				t.Fatalf("Results weren't the same. %d != %d for index %d,%d",
					result.At(x, y),
					expectedResult.At(x, y),
					x,
					y,
				)
			}

		}
	}
}
