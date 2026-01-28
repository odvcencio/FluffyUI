package graphics

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"github.com/odvcencio/fluffyui/backend"
)

const kittyChunkSize = 4096

func buildRGBAImage(pixels *PixelBuffer, cellWidth, cellHeight int, protocol backend.ImageProtocol) backend.Image {
	if pixels == nil {
		return backend.Image{}
	}
	w, h := pixels.Size()
	if w <= 0 || h <= 0 {
		return backend.Image{}
	}
	data := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := pixels.Get(x, y)
			r, g, b, a := pixelToRGBA(p)
			idx := (y*w + x) * 4
			data[idx] = r
			data[idx+1] = g
			data[idx+2] = b
			data[idx+3] = a
		}
	}
	return backend.Image{
		Width:      w,
		Height:     h,
		CellWidth:  cellWidth,
		CellHeight: cellHeight,
		Format:     backend.ImageFormatRGBA,
		Protocol:   protocol,
		Pixels:     data,
	}
}

func pixelToRGBA(p Pixel) (uint8, uint8, uint8, uint8) {
	if !p.Set || p.Color == backend.ColorDefault {
		return 0, 0, 0, 0
	}
	alpha := clampFloat64(float64(p.Alpha), 0, 1)
	if alpha <= 0 {
		return 0, 0, 0, 0
	}
	var r, g, b uint8
	if p.Color.IsRGB() {
		r, g, b = p.Color.RGB()
	} else {
		r, g, b = paletteIndexToRGB(int(p.Color))
	}
	return r, g, b, uint8(alpha * 255)
}

func clampFloat64(v, minValue, maxValue float64) float64 {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func paletteIndexToRGB(index int) (uint8, uint8, uint8) {
	if index < 0 {
		return 0, 0, 0
	}
	if index < len(ansi16) {
		c := ansi16[index]
		return c[0], c[1], c[2]
	}
	if index >= 16 && index <= 231 {
		idx := index - 16
		r := idx / 36
		g := (idx / 6) % 6
		b := idx % 6
		return cube6[r], cube6[g], cube6[b]
	}
	if index >= 232 && index <= 255 {
		v := uint8(8 + (index-232)*10)
		return v, v, v
	}
	return 0, 0, 0
}

var ansi16 = [16][3]uint8{
	{0, 0, 0},
	{128, 0, 0},
	{0, 128, 0},
	{128, 128, 0},
	{0, 0, 128},
	{128, 0, 128},
	{0, 128, 128},
	{192, 192, 192},
	{128, 128, 128},
	{255, 0, 0},
	{0, 255, 0},
	{255, 255, 0},
	{0, 0, 255},
	{255, 0, 255},
	{0, 255, 255},
	{255, 255, 255},
}

var cube6 = [6]uint8{0, 95, 135, 175, 215, 255}

// EncodeKitty returns kitty graphics payload bytes for the image.
func EncodeKitty(img backend.Image) []byte {
	if img.Format != backend.ImageFormatRGBA || img.Width <= 0 || img.Height <= 0 {
		return nil
	}
	payload := base64.StdEncoding.EncodeToString(img.Pixels)
	params := fmt.Sprintf("a=T,f=32,s=%d,v=%d,c=%d,r=%d,q=2,t=d", img.Width, img.Height, img.CellWidth, img.CellHeight)
	var out bytes.Buffer
	for i := 0; i < len(payload); i += kittyChunkSize {
		end := i + kittyChunkSize
		if end > len(payload) {
			end = len(payload)
		}
		more := 0
		if end < len(payload) {
			more = 1
		}
		if i == 0 {
			out.WriteString("\x1b_G")
			out.WriteString(params)
			out.WriteString(",m=")
			out.WriteString(strconv.Itoa(more))
			out.WriteString(";")
		} else {
			out.WriteString("\x1b_Gm=")
			out.WriteString(strconv.Itoa(more))
			out.WriteString(";")
		}
		out.WriteString(payload[i:end])
		out.WriteString("\x1b\\")
	}
	return out.Bytes()
}

// EncodeSixel returns sixel payload bytes for the image.
func EncodeSixel(img backend.Image) []byte {
	if img.Format != backend.ImageFormatRGBA || img.Width <= 0 || img.Height <= 0 {
		return nil
	}
	w := img.Width
	h := img.Height
	pixels := img.Pixels
	indices := make([]int, w*h)
	used := make(map[int]struct{})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			r := pixels[idx]
			g := pixels[idx+1]
			b := pixels[idx+2]
			a := pixels[idx+3]
			if a == 0 {
				indices[y*w+x] = -1
				continue
			}
			colorIndex := rgbToSixelIndex(r, g, b)
			indices[y*w+x] = colorIndex
			used[colorIndex] = struct{}{}
		}
	}
	palette := make([]int, 0, len(used))
	for index := range used {
		palette = append(palette, index)
	}
	sort.Ints(palette)
	var out bytes.Buffer
	out.WriteString("\x1bPq")
	for _, index := range palette {
		r, g, b := paletteIndexToRGB(index)
		out.WriteString("#")
		out.WriteString(strconv.Itoa(index))
		out.WriteString(";2;")
		out.WriteString(strconv.Itoa(int(r) * 100 / 255))
		out.WriteString(";")
		out.WriteString(strconv.Itoa(int(g) * 100 / 255))
		out.WriteString(";")
		out.WriteString(strconv.Itoa(int(b) * 100 / 255))
	}
	for y0 := 0; y0 < h; y0 += 6 {
		for _, index := range palette {
			line := make([]byte, w)
			hasBits := false
			for x := 0; x < w; x++ {
				mask := 0
				for bit := 0; bit < 6; bit++ {
					y := y0 + bit
					if y >= h {
						continue
					}
					if indices[y*w+x] == index {
						mask |= 1 << bit
					}
				}
				if mask != 0 {
					hasBits = true
				}
				line[x] = byte(mask + 63)
			}
			if !hasBits {
				continue
			}
			out.WriteString("#")
			out.WriteString(strconv.Itoa(index))
			out.Write(line)
			out.WriteByte('$')
		}
		if y0+6 < h {
			out.WriteByte('-')
		}
	}
	out.WriteString("\x1b\\")
	return out.Bytes()
}

func rgbToSixelIndex(r, g, b uint8) int {
	r6 := int(r) * 5 / 255
	g6 := int(g) * 5 / 255
	b6 := int(b) * 5 / 255
	return 16 + 36*r6 + 6*g6 + b6
}
