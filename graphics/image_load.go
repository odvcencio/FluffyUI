package graphics

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// LoadImage loads an image from disk using any registered decoder.
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", path, err)
	}
	defer file.Close()
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}
	_ = format // available if needed for debugging
	return img, nil
}

// LoadImageScaled loads an image and scales it to fit within max dimensions.
func LoadImageScaled(path string, maxWidth, maxHeight int) (image.Image, error) {
	img, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return ScaleImage(img, maxWidth, maxHeight), nil
}

// ScaleImage scales an image to fit within max dimensions using nearest neighbor.
func ScaleImage(img image.Image, maxWidth, maxHeight int) image.Image {
	if img == nil {
		return nil
	}
	if maxWidth <= 0 || maxHeight <= 0 {
		return img
	}
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return img
	}
	scale := math.Min(float64(maxWidth)/float64(srcW), float64(maxHeight)/float64(srcH))
	if scale >= 1 {
		return img
	}
	targetW := int(math.Round(float64(srcW) * scale))
	targetH := int(math.Round(float64(srcH) * scale))
	if targetW <= 0 {
		targetW = 1
	}
	if targetH <= 0 {
		targetH = 1
	}
	return resizeNearestNeighbor(img, targetW, targetH)
}

// ScaleMode determines the interpolation method for scaling.
type ScaleMode int

const (
	// ScaleNearest uses nearest-neighbor interpolation (fast, pixelated).
	ScaleNearest ScaleMode = iota
	// ScaleBilinear uses bilinear interpolation (smoother, slower).
	ScaleBilinear
)

// ScaleImageMode scales an image using the specified interpolation mode.
func ScaleImageMode(img image.Image, maxWidth, maxHeight int, mode ScaleMode) image.Image {
	if img == nil || maxWidth <= 0 || maxHeight <= 0 {
		return img
	}
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return img
	}
	scale := math.Min(float64(maxWidth)/float64(srcW), float64(maxHeight)/float64(srcH))
	if scale >= 1 {
		return img
	}
	targetW := int(math.Round(float64(srcW) * scale))
	targetH := int(math.Round(float64(srcH) * scale))
	if targetW <= 0 {
		targetW = 1
	}
	if targetH <= 0 {
		targetH = 1
	}
	switch mode {
	case ScaleBilinear:
		return resizeBilinear(img, targetW, targetH)
	default:
		return resizeNearestNeighbor(img, targetW, targetH)
	}
}

func resizeNearestNeighbor(img image.Image, targetW, targetH int) image.Image {
	if img == nil || targetW <= 0 || targetH <= 0 {
		return img
	}
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	for y := 0; y < targetH; y++ {
		srcY := srcBounds.Min.Y + y*srcH/targetH
		for x := 0; x < targetW; x++ {
			srcX := srcBounds.Min.X + x*srcW/targetW
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}

func resizeBilinear(img image.Image, targetW, targetH int) image.Image {
	if img == nil || targetW <= 0 || targetH <= 0 {
		return img
	}
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	xRatio := float64(srcW) / float64(targetW)
	yRatio := float64(srcH) / float64(targetH)

	for y := 0; y < targetH; y++ {
		srcYf := float64(y) * yRatio
		srcY := int(srcYf)
		yFrac := srcYf - float64(srcY)
		srcY += srcBounds.Min.Y
		srcY1 := srcY + 1
		if srcY1 >= srcBounds.Max.Y {
			srcY1 = srcBounds.Max.Y - 1
		}

		for x := 0; x < targetW; x++ {
			srcXf := float64(x) * xRatio
			srcX := int(srcXf)
			xFrac := srcXf - float64(srcX)
			srcX += srcBounds.Min.X
			srcX1 := srcX + 1
			if srcX1 >= srcBounds.Max.X {
				srcX1 = srcBounds.Max.X - 1
			}

			// Sample four corners
			c00 := img.At(srcX, srcY)
			c10 := img.At(srcX1, srcY)
			c01 := img.At(srcX, srcY1)
			c11 := img.At(srcX1, srcY1)

			// Interpolate
			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			r := bilinearInterp(r00, r10, r01, r11, xFrac, yFrac)
			g := bilinearInterp(g00, g10, g01, g11, xFrac, yFrac)
			b := bilinearInterp(b00, b10, b01, b11, xFrac, yFrac)
			a := bilinearInterp(a00, a10, a01, a11, xFrac, yFrac)

			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func bilinearInterp(c00, c10, c01, c11 uint32, xFrac, yFrac float64) uint32 {
	top := float64(c00)*(1-xFrac) + float64(c10)*xFrac
	bot := float64(c01)*(1-xFrac) + float64(c11)*xFrac
	return uint32(top*(1-yFrac) + bot*yFrac)
}
