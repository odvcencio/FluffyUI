//go:build !js

package gpu

import (
	"unsafe"

	"github.com/odvcencio/fluffyui/gpu/opengl/gl"
)

func applyLayout(layout *VertexLayout) {
	if layout == nil {
		return
	}
	stride := layout.Stride
	if stride <= 0 {
		stride = inferLayoutStride(layout)
	}
	for _, attr := range layout.Attributes {
		if attr.Size <= 0 {
			continue
		}
		attrStride := attr.Stride
		if attrStride <= 0 {
			attrStride = stride
		}
		normalized := uint8(gl.FALSE)
		if attr.Normalized {
			normalized = uint8(gl.TRUE)
		}
		gl.EnableVertexAttribArray(attr.Index)
		gl.VertexAttribPointer(attr.Index, attr.Size, attr.Type, normalized, int32(attrStride), unsafe.Pointer(uintptr(attr.Offset)))
	}
}

func inferLayoutStride(layout *VertexLayout) int {
	max := 0
	for _, attr := range layout.Attributes {
		end := attr.Offset + int(attr.Size)*attributeTypeSize(attr.Type)
		if end > max {
			max = end
		}
	}
	return max
}

func attributeTypeSize(typ uint32) int {
	switch typ {
	case gl.UNSIGNED_SHORT:
		return 2
	default:
		return 4
	}
}
