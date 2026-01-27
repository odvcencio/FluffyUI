//go:build darwin

package gpu

// Metal constants and structs (minimal subset).
const (
	mtlPixelFormatRGBA8Unorm = 70

	mtlTextureUsageShaderRead   = 1 << 0
	mtlTextureUsageShaderWrite  = 1 << 1
	mtlTextureUsageRenderTarget = 1 << 2

	mtlStorageModeShared = 0

	mtlLoadActionDontCare = 0
	mtlLoadActionLoad     = 1
	mtlLoadActionClear    = 2

	mtlStoreActionDontCare = 0
	mtlStoreActionStore    = 1

	mtlPrimitiveTypeTriangle = 3

	mtlIndexTypeUInt16 = 0

	mtlVertexFormatFloat2 = 29
	mtlVertexFormatFloat4 = 31

	mtlVertexStepFunctionPerVertex = 1

	mtlBlendFactorZero                = 0
	mtlBlendFactorOne                 = 1
	mtlBlendFactorSourceAlpha         = 4
	mtlBlendFactorOneMinusSourceAlpha = 5

	mtlBlendOperationAdd = 0
)

type mtlOrigin struct {
	X uint64
	Y uint64
	Z uint64
}

type mtlSize struct {
	Width  uint64
	Height uint64
	Depth  uint64
}

type mtlRegion struct {
	Origin mtlOrigin
	Size   mtlSize
}

type mtlScissorRect struct {
	X      uint64
	Y      uint64
	Width  uint64
	Height uint64
}

type mtlViewport struct {
	OriginX float64
	OriginY float64
	Width   float64
	Height  float64
	ZNear   float64
	ZFar    float64
}

type mtlClearColor struct {
	R float64
	G float64
	B float64
	A float64
}
