package hmap

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path"

	"strings"

	"github.com/nfnt/resize"
)

// Color the color to use for highlighting sectors
type Color string

const (
	// Green - green
	Green Color = Color("green")

	// Orange - orange
	Orange Color = Color("orange")

	// Pink - pink
	Pink Color = Color("pink")

	// Yellow - yellow
	Yellow Color = Color("yellow")

	// Red - red
	Red Color = Color("red")
	// Warn - warning pattern (yellow-black stripes)
	Warn Color = Color("warn")
)

// DefaultColor the default highlight color
const DefaultColor = Green

type hex struct {
	img  image.Image
	rect image.Rectangle
}

const cellSizeHight = (1.0 / 7.0)         // 7 cells in a map horizontally
const cellSizeWight = (1.0 / 7.0) - 0.007 // 7 cells in a map vertically including some offset

// GenerateBaseImage generates the base image, composed of the real in game map with overlayed coordinates.
func GenerateBaseImage(layers []string) (draw.Image, error) {
	var baseImage *image.RGBA

	var bounds image.Rectangle
	for i, layer := range layers {
		img, err := LoadImage(layer)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			bounds = img.Bounds()
			baseImage = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		}

		if bounds != img.Bounds() {
			img = resize.Resize(uint(bounds.Dx()), uint(bounds.Dy()), img, resize.Lanczos3)
		}

		draw.Draw(baseImage, baseImage.Bounds(), img, image.Point{0, 0}, draw.Over)
	}

	return baseImage, nil
}

// HighlightCoord highlights the provided coordinate.
func HighlightCoord(baseImage draw.Image, coords string, color Color) (draw.Image, error) {
	hex, err := getHex(coords, color, baseImage.Bounds())
	if err != nil {
		return nil, err
	}

	draw.Draw(baseImage, hex.rect, hex.img, image.Point{0, 0}, draw.Over)
	return baseImage, nil
}

// LoadImage loads an image from the disc. Supported formats are .jpeg and .png
func LoadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var img image.Image
	if strings.HasSuffix(filePath, ".jpeg") {
		img, err = jpeg.Decode(file)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filePath, ".png") {
		img, err = png.Decode(file)
		if err != nil {
			return nil, err
		}
	}

	return img, nil
}

// getHex constructs and returns the hex object, containing the image and to rectangle to use on top of the provided base image bounds
func getHex(coord string, color Color, baseBounds image.Rectangle) (hex, error) {
	if !isValidCoord(coord) {
		return hex{}, fmt.Errorf("invalid coordinate: %s", coord)
	}
	hexImg, err := LoadImage(path.Join(staticPath, fmt.Sprintf("hex_%s.png", color)))
	if err != nil {
		return hex{}, err
	}

	hexImageResized := resize.Resize(0, uint(cellSizeHight*float64(baseBounds.Dy())), hexImg, resize.Lanczos3)
	hexRect := getTargetPoint(coord, baseBounds, hexImageResized.Bounds())

	return hex{img: hexImageResized, rect: hexRect}, nil
}

// isValidCoord checks if the provided coordinate string is valid for the current map schema
func isValidCoord(coord string) bool {
	directions := []string{
		"a1", "a2", "a3", "a4",
		"b1", "b2", "b3", "b4", "b5",
		"c1", "c2", "c3", "c4", "c5", "c6",
		"d1", "d2", "d3", "d4", "d5", "d6", "d7",
		"e1", "e2", "e3", "e4", "e5", "e6",
		"f1", "f2", "f3", "f4", "f5",
		"g1", "g2", "g3", "g4",
	}

	coord = strings.ToLower(coord)
	for _, c := range directions {
		if coord == c {
			return true
		}
	}
	return false
}

// getTargetPoint calculates the place where to put the hex rectangle in the base image
func getTargetPoint(coords string, base, hex image.Rectangle) image.Rectangle {
	coordPoint := image.Point{
		int(float64(base.Dx()) * newCellPoint(coords).x),
		int(float64(base.Dy()) * newCellPoint(coords).y),
	}

	hexRect := image.Rectangle{coordPoint, coordPoint.Add(hex.Max)}

	return hexRect
}
