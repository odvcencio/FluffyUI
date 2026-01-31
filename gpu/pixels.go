package gpu

import (
	"image"
	"image/color"
)

func rgbaFromFloats(r, g, b, a float32) color.RGBA {
	return color.RGBA{
		R: floatToByte(r),
		G: floatToByte(g),
		B: floatToByte(b),
		A: floatToByte(a),
	}
}

func floatToByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func clearPixels(pixels []byte, w, h int, col color.RGBA) {
	if len(pixels) == 0 || w <= 0 || h <= 0 {
		return
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			pixels[idx] = col.R
			pixels[idx+1] = col.G
			pixels[idx+2] = col.B
			pixels[idx+3] = col.A
		}
	}
}

func blendPixel(pixels []byte, w, h, x, y int, col color.RGBA) {
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	idx := (y*w + x) * 4
	dr := pixels[idx]
	dg := pixels[idx+1]
	db := pixels[idx+2]
	da := pixels[idx+3]
	sr := col.R
	sg := col.G
	sb := col.B
	sa := col.A
	outR, outG, outB, outA := blendRGBA(dr, dg, db, da, sr, sg, sb, sa)
	pixels[idx] = outR
	pixels[idx+1] = outG
	pixels[idx+2] = outB
	pixels[idx+3] = outA
}

func setPixel(pixels []byte, w, h, x, y int, col color.RGBA) {
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	idx := (y*w + x) * 4
	pixels[idx] = col.R
	pixels[idx+1] = col.G
	pixels[idx+2] = col.B
	pixels[idx+3] = col.A
}

func blendRGBA(dr, dg, db, da, sr, sg, sb, sa uint8) (uint8, uint8, uint8, uint8) {
	srcA := float32(sa) / 255
	dstA := float32(da) / 255
	outA := srcA + dstA*(1-srcA)
	if outA <= 0 {
		return 0, 0, 0, 0
	}
	srcR := float32(sr) / 255
	srcG := float32(sg) / 255
	srcB := float32(sb) / 255
	dstR := float32(dr) / 255
	dstG := float32(dg) / 255
	dstB := float32(db) / 255
	outR := (srcR*srcA + dstR*dstA*(1-srcA)) / outA
	outG := (srcG*srcA + dstG*dstA*(1-srcA)) / outA
	outB := (srcB*srcA + dstB*dstA*(1-srcA)) / outA
	return floatToByte(outR), floatToByte(outG), floatToByte(outB), floatToByte(outA)
}

func scalePixels(src []byte, srcW, srcH, dstW, dstH int) []byte {
	if dstW <= 0 || dstH <= 0 || srcW <= 0 || srcH <= 0 {
		return nil
	}
	dst := make([]byte, dstW*dstH*4)
	for y := 0; y < dstH; y++ {
		sy := y * srcH / dstH
		for x := 0; x < dstW; x++ {
			sx := x * srcW / dstW
			srcIdx := (sy*srcW + sx) * 4
			dstIdx := (y*dstW + x) * 4
			copy(dst[dstIdx:dstIdx+4], src[srcIdx:srcIdx+4])
		}
	}
	return dst
}

func cropPixels(src []byte, srcW, srcH int, rect image.Rectangle) []byte {
	if rect.Empty() {
		rect = image.Rect(0, 0, srcW, srcH)
	}
	if rect.Min.X < 0 {
		rect.Min.X = 0
	}
	if rect.Min.Y < 0 {
		rect.Min.Y = 0
	}
	if rect.Max.X > srcW {
		rect.Max.X = srcW
	}
	if rect.Max.Y > srcH {
		rect.Max.Y = srcH
	}
	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return nil
	}
	out := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcIdx := ((rect.Min.Y+y)*srcW + (rect.Min.X + x)) * 4
			dstIdx := (y*w + x) * 4
			copy(out[dstIdx:dstIdx+4], src[srcIdx:srcIdx+4])
		}
	}
	return out
}

func flipPixelsVertical(pixels []byte, w, h int) {
	if len(pixels) == 0 || w <= 0 || h <= 1 {
		return
	}
	stride := w * 4
	tmp := make([]byte, stride)
	top := 0
	bottom := h - 1
	for top < bottom {
		topIdx := top * stride
		bottomIdx := bottom * stride
		copy(tmp, pixels[topIdx:topIdx+stride])
		copy(pixels[topIdx:topIdx+stride], pixels[bottomIdx:bottomIdx+stride])
		copy(pixels[bottomIdx:bottomIdx+stride], tmp)
		top++
		bottom--
	}
}
