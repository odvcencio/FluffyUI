//go:build darwin

package gpu

import (
	"image"
	"image/color"
	"runtime"
	"testing"
)

// Test helper functions that don't require a GPU

func TestParseMetalEntryPoints(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		wantVertex     string
		wantFragment   string
	}{
		{
			name:         "default names when no definitions",
			source:       "// some code\n",
			wantVertex:   "vertex_main",
			wantFragment: "fragment_main",
		},
		{
			name: "parse custom vertex function",
			source: `vertex VertexOut myVertex(VertexIn in [[stage_in]]) {
				return out;
			}`,
			wantVertex:   "myVertex",
			wantFragment: "fragment_main",
		},
		{
			name: "parse custom fragment function",
			source: `fragment float4 myFragment(VertexOut in [[stage_in]]) {
				return float4(1);
			}`,
			wantVertex:   "vertex_main",
			wantFragment: "myFragment",
		},
		{
			name: "parse both custom functions",
			source: `vertex VertexOut customVert(VertexIn in [[stage_in]]) {
				VertexOut out;
				return out;
			}
			fragment float4 customFrag(VertexOut in [[stage_in]]) {
				return float4(1);
			}`,
			wantVertex:   "customVert",
			wantFragment: "customFrag",
		},
		{
			name: "handle function with parenthesis in name",
			source: `vertex VertexOut myFunc(VertexIn in) {
			}`,
			wantVertex:   "myFunc",
			wantFragment: "fragment_main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVertex, gotFragment := parseMetalEntryPoints(tt.source)
			if gotVertex != tt.wantVertex {
				t.Errorf("vertex = %q, want %q", gotVertex, tt.wantVertex)
			}
			if gotFragment != tt.wantFragment {
				t.Errorf("fragment = %q, want %q", gotFragment, tt.wantFragment)
			}
		})
	}
}

func TestParseMetalVertexLayout(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		wantStride     int
		wantAttr1Fmt   int
		wantAttr1Off   int
	}{
		{
			name:         "default layout when no attribute(1)",
			source:       "// no attributes",
			wantStride:   16,
			wantAttr1Fmt: mtlVertexFormatFloat2,
			wantAttr1Off: 8,
		},
		{
			name:         "float2 attribute(1)",
			source:       "float2 texcoord [[attribute(1)]]",
			wantStride:   16,
			wantAttr1Fmt: mtlVertexFormatFloat2,
			wantAttr1Off: 8,
		},
		{
			name:         "float4 attribute(1)",
			source:       "float4 color [[attribute(1)]]",
			wantStride:   24,
			wantAttr1Fmt: mtlVertexFormatFloat4,
			wantAttr1Off: 8,
		},
		{
			name: "float4 with whitespace",
			source: `struct VertexIn {
				float2 pos [[attribute(0)]];
				float4 col [[attribute(1)]];
			};`,
			wantStride:   24,
			wantAttr1Fmt: mtlVertexFormatFloat4,
			wantAttr1Off: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := parseMetalVertexLayout(tt.source)
			if layout.stride != tt.wantStride {
				t.Errorf("stride = %d, want %d", layout.stride, tt.wantStride)
			}
			if layout.attr1Fmt != tt.wantAttr1Fmt {
				t.Errorf("attr1Fmt = %d, want %d", layout.attr1Fmt, tt.wantAttr1Fmt)
			}
			if layout.attr1Offset != tt.wantAttr1Off {
				t.Errorf("attr1Offset = %d, want %d", layout.attr1Offset, tt.wantAttr1Off)
			}
		})
	}
}

func TestParseMetalUniforms(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantLen  int
		wantName string
		wantKind metalUniformKind
		wantIdx  int
	}{
		{
			name:    "no uniforms",
			source:  "// no uniforms",
			wantLen: 0,
		},
		{
			name:     "float uniform",
			source:   "constant float& uAlpha [[buffer(2)]]",
			wantLen:  1,
			wantName: "uAlpha",
			wantKind: metalUniformFloat,
			wantIdx:  2,
		},
		{
			name:     "float2 uniform",
			source:   "constant float2& uOffset [[buffer(1)]]",
			wantLen:  1,
			wantName: "uOffset",
			wantKind: metalUniformFloat2,
			wantIdx:  1,
		},
		{
			name:     "float4 uniform",
			source:   "constant float4& uColor [[buffer(3)]]",
			wantLen:  1,
			wantName: "uColor",
			wantKind: metalUniformFloat4,
			wantIdx:  3,
		},
		{
			name:     "float4x4 uniform",
			source:   "constant float4x4& uTransform [[buffer(1)]]",
			wantLen:  1,
			wantName: "uTransform",
			wantKind: metalUniformFloat4x4,
			wantIdx:  1,
		},
		{
			name: "multiple uniforms",
			source: `constant float4x4& uTransform [[buffer(1)]]
			         constant float& uAlpha [[buffer(2)]]`,
			wantLen:  2,
			wantName: "uTransform",
			wantKind: metalUniformFloat4x4,
			wantIdx:  1,
		},
		{
			name:    "unsupported type ignored",
			source:  "constant int& uCount [[buffer(1)]]",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniforms := parseMetalUniforms(tt.source)
			if len(uniforms) != tt.wantLen {
				t.Errorf("len(uniforms) = %d, want %d", len(uniforms), tt.wantLen)
				return
			}
			if tt.wantLen > 0 && tt.wantName != "" {
				meta, ok := uniforms[tt.wantName]
				if !ok {
					t.Errorf("uniform %q not found", tt.wantName)
					return
				}
				if meta.kind != tt.wantKind {
					t.Errorf("kind = %d, want %d", meta.kind, tt.wantKind)
				}
				if meta.index != tt.wantIdx {
					t.Errorf("index = %d, want %d", meta.index, tt.wantIdx)
				}
			}
		})
	}
}

func TestFloatFromAny(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    float32
		wantOK  bool
	}{
		{"float32", float32(1.5), 1.5, true},
		{"float64", float64(2.5), 2.5, true},
		{"int", 3, 3.0, true},
		{"int32", int32(4), 4.0, true},
		{"string", "1.0", 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := floatFromAny(tt.value)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("value = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntFromAny(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		want   int
		wantOK bool
	}{
		{"int", 42, 42, true},
		{"int32", int32(24), 24, true},
		{"uint32", uint32(12), 12, true},
		{"string", "42", 0, false},
		{"float32", float32(1.5), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intFromAny(tt.value)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("value = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec2FromAny(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		want   [2]float32
		wantOK bool
	}{
		{"[2]float32", [2]float32{1, 2}, [2]float32{1, 2}, true},
		{"[2]float64", [2]float64{3, 4}, [2]float32{3, 4}, true},
		{"[]float32 sufficient", []float32{5, 6, 7}, [2]float32{5, 6}, true},
		{"[]float64 sufficient", []float64{8, 9}, [2]float32{8, 9}, true},
		{"[]float32 insufficient", []float32{1}, [2]float32{}, false},
		{"string", "1,2", [2]float32{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := vec2FromAny(tt.value)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("value = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec4FromAny(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		want   [4]float32
		wantOK bool
	}{
		{"[4]float32", [4]float32{1, 2, 3, 4}, [4]float32{1, 2, 3, 4}, true},
		{"[4]float64", [4]float64{5, 6, 7, 8}, [4]float32{5, 6, 7, 8}, true},
		{"[]float32 sufficient", []float32{1, 2, 3, 4, 5}, [4]float32{1, 2, 3, 4}, true},
		{"[]float64 sufficient", []float64{9, 10, 11, 12}, [4]float32{9, 10, 11, 12}, true},
		{"[]float32 insufficient", []float32{1, 2, 3}, [4]float32{}, false},
		{"color.RGBA", color.RGBA{R: 255, G: 128, B: 64, A: 255}, [4]float32{1, 128.0 / 255, 64.0 / 255, 1}, true},
		{"string", "1,2,3,4", [4]float32{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := vec4FromAny(tt.value)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				// Use approximate comparison for color conversion
				for i := 0; i < 4; i++ {
					diff := got[i] - tt.want[i]
					if diff < -0.01 || diff > 0.01 {
						t.Errorf("value[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestMat4FromAny(t *testing.T) {
	identity16 := [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}

	tests := []struct {
		name   string
		value  any
		wantOK bool
	}{
		{"[16]float32", identity16, true},
		{"[]float32 sufficient", identity16[:], true},
		{"[]float32 insufficient", []float32{1, 2, 3}, false},
		{"Matrix3", Identity(), true},
		{"[9]float32", [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}, true},
		{"string", "matrix", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := mat4FromAny(tt.value)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestMatrix3ToMetal(t *testing.T) {
	m := Identity()
	result := matrix3ToMetal(m)

	// Check that identity matrix converts correctly
	// Should be: a=1, b=0, c=0, d=0, e=1, f=0
	// Result should be column-major 4x4
	expected := [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	for i := 0; i < 16; i++ {
		if result[i] != expected[i] {
			t.Errorf("matrix[%d] = %v, want %v", i, result[i], expected[i])
		}
	}
}

func TestToMetalUniformValue(t *testing.T) {
	tests := []struct {
		name   string
		kind   metalUniformKind
		value  any
		wantNil bool
	}{
		{"float valid", metalUniformFloat, float32(1.0), false},
		{"float invalid", metalUniformFloat, "string", true},
		{"float2 valid", metalUniformFloat2, [2]float32{1, 2}, false},
		{"float2 invalid", metalUniformFloat2, float32(1.0), true},
		{"float4 valid", metalUniformFloat4, [4]float32{1, 2, 3, 4}, false},
		{"float4 invalid", metalUniformFloat4, [2]float32{1, 2}, true},
		{"float4x4 valid", metalUniformFloat4x4, [16]float32{}, false},
		{"float4x4 invalid", metalUniformFloat4x4, [4]float32{}, true},
		{"int valid", metalUniformInt, 42, false},
		{"int invalid", metalUniformInt, "42", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := toMetalUniformValue(tt.kind, tt.value)
			if (val == nil) != tt.wantNil {
				t.Errorf("result nil = %v, want nil = %v", val == nil, tt.wantNil)
			}
		})
	}
}

func TestMetalUniformBytes(t *testing.T) {
	tests := []struct {
		name     string
		kind     metalUniformKind
		wantSize uintptr
	}{
		{"float", metalUniformFloat, 4},
		{"float2", metalUniformFloat2, 8},
		{"float4", metalUniformFloat4, 16},
		{"float4x4", metalUniformFloat4x4, 64},
		{"int", metalUniformInt, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := &metalUniformValue{kind: tt.kind}
			meta := metalUniformMeta{kind: tt.kind, size: tt.wantSize}
			ptr, size := metalUniformBytes(val, meta)
			if ptr == nil {
				t.Error("ptr is nil")
			}
			if size != tt.wantSize {
				t.Errorf("size = %d, want %d", size, tt.wantSize)
			}
		})
	}

	// Test nil value
	ptr, size := metalUniformBytes(nil, metalUniformMeta{})
	if ptr != nil || size != 0 {
		t.Error("expected nil and 0 for nil value")
	}

	// Test unknown kind
	val := &metalUniformValue{kind: metalUniformKind(99)}
	meta := metalUniformMeta{kind: metalUniformKind(99), size: 0}
	ptr, size = metalUniformBytes(val, meta)
	if ptr != nil || size != 0 {
		t.Error("expected nil and 0 for unknown kind")
	}
}

// skipIfNoMetal skips the test if Metal is not available.
func skipIfNoMetal(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("Metal tests only run on macOS")
	}
	// Try to initialize objc - if this fails, Metal is not available
	if err := objcInit(); err != nil {
		t.Skipf("Metal not available: %v", err)
	}
}

// GPU-dependent tests - these will skip if no GPU is available

func TestMetalDriverCreation(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed (no GPU?): %v", err)
	}
	defer driver.Dispose()

	if driver.Backend() != BackendMetal {
		t.Errorf("Backend() = %v, want BackendMetal", driver.Backend())
	}

	if err := driver.Init(); err != nil {
		t.Errorf("Init() error = %v", err)
	}

	if driver.MaxTextureSize() <= 0 {
		t.Errorf("MaxTextureSize() = %d, want > 0", driver.MaxTextureSize())
	}
}

func TestMetalDriverCreationOnNonDarwin(t *testing.T) {
	// This test verifies the behavior when runtime.GOOS != "darwin"
	// On darwin, this is already tested by the build constraint
	// But we can test the nil driver behavior
	var nilDriver *metalDriver
	if nilDriver.Backend() != BackendMetal {
		t.Errorf("nil driver Backend() should still return BackendMetal")
	}
	if nilDriver.MaxTextureSize() != 0 {
		t.Errorf("nil driver MaxTextureSize() should return 0")
	}
}

func TestMetalTextureCreation(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid 64x64", 64, 64, false},
		{"valid 1x1", 1, 1, false},
		{"valid 256x128", 256, 128, false},
		{"invalid 0x0", 0, 0, true},
		{"invalid -1x1", -1, 1, true},
		{"invalid 1x-1", 1, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tex, err := driver.NewTexture(tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTexture() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				w, h := tex.Size()
				if w != tt.width || h != tt.height {
					t.Errorf("Size() = (%d, %d), want (%d, %d)", w, h, tt.width, tt.height)
				}
				// Texture ID for Metal is always 0 (internal handle)
				if tex.ID() != 0 {
					t.Errorf("ID() = %d, want 0 for Metal texture", tex.ID())
				}
				tex.Dispose()
			}
		})
	}
}

func TestMetalTextureUpload(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	tex, err := driver.NewTexture(4, 4)
	if err != nil {
		t.Fatalf("NewTexture() error = %v", err)
	}
	defer tex.Dispose()

	// Create red pixels
	pixels := make([]byte, 4*4*4)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = 255   // R
		pixels[i+1] = 0   // G
		pixels[i+2] = 0   // B
		pixels[i+3] = 255 // A
	}

	// Test full upload
	tex.Upload(pixels, image.Rectangle{})

	// Test partial upload
	tex.Upload(pixels[:16], image.Rect(0, 0, 2, 2))

	// Test upload with invalid region (should clamp)
	tex.Upload(pixels, image.Rect(-1, -1, 10, 10))

	// Test upload with empty pixels (should be no-op)
	tex.Upload(nil, image.Rectangle{})
	tex.Upload([]byte{}, image.Rectangle{})
}

func TestMetalFramebufferCreation(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid 64x64", 64, 64, false},
		{"valid 1x1", 1, 1, false},
		{"invalid 0x0", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := driver.NewFramebuffer(tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFramebuffer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				w, h := fb.Size()
				if w != tt.width || h != tt.height {
					t.Errorf("Size() = (%d, %d), want (%d, %d)", w, h, tt.width, tt.height)
				}
				// Framebuffer ID for Metal is always 0
				if fb.ID() != 0 {
					t.Errorf("ID() = %d, want 0 for Metal framebuffer", fb.ID())
				}
				// Should have a texture
				if fb.Texture() == nil {
					t.Error("Texture() returned nil")
				}
				fb.Bind()
				fb.Dispose()
			}
		})
	}
}

func TestMetalShaderCompilation(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	// Load solid shader source
	src, err := LoadShaderSource("solid")
	if err != nil {
		t.Fatalf("LoadShaderSource() error = %v", err)
	}

	shader, err := driver.NewShader(src)
	if err != nil {
		t.Fatalf("NewShader() error = %v", err)
	}
	defer shader.Dispose()

	// Shader ID for Metal is always 0
	if shader.ID() != 0 {
		t.Errorf("ID() = %d, want 0 for Metal shader", shader.ID())
	}
}

func TestMetalShaderUniformSetting(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	src, err := LoadShaderSource("solid")
	if err != nil {
		t.Fatalf("LoadShaderSource() error = %v", err)
	}

	shader, err := driver.NewShader(src)
	if err != nil {
		t.Fatalf("NewShader() error = %v", err)
	}
	defer shader.Dispose()

	// Test setting various uniform types
	shader.SetUniform("uTransform", Identity())
	shader.SetUniform("uTransform", [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1})

	// Test setting non-existent uniform (should be ignored)
	shader.SetUniform("nonexistent", float32(1.0))

	// Test setting with empty name (should be ignored)
	shader.SetUniform("", float32(1.0))
}

func TestMetalShaderCompilationErrors(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	tests := []struct {
		name    string
		src     ShaderSource
		wantErr bool
	}{
		{
			name:    "empty metal source",
			src:     ShaderSource{Metal: ""},
			wantErr: true,
		},
		{
			name:    "invalid metal source",
			src:     ShaderSource{Metal: "this is not valid metal code"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shader, err := driver.NewShader(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewShader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if shader != nil {
				shader.Dispose()
			}
		})
	}
}

func TestMetalClearOperation(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	fb, err := driver.NewFramebuffer(4, 4)
	if err != nil {
		t.Fatalf("NewFramebuffer() error = %v", err)
	}
	defer fb.Dispose()

	fb.Bind()

	// Clear to red
	driver.Clear(1.0, 0.0, 0.0, 1.0)

	// Read pixels back
	pixels, err := driver.ReadPixels(fb, image.Rectangle{})
	if err != nil {
		t.Fatalf("ReadPixels() error = %v", err)
	}

	// Check that all pixels are red
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] != 255 || pixels[i+1] != 0 || pixels[i+2] != 0 || pixels[i+3] != 255 {
			t.Errorf("pixel %d: got RGBA(%d,%d,%d,%d), want (255,0,0,255)",
				i/4, pixels[i], pixels[i+1], pixels[i+2], pixels[i+3])
			break
		}
	}
}

func TestMetalReadPixels(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	fb, err := driver.NewFramebuffer(8, 8)
	if err != nil {
		t.Fatalf("NewFramebuffer() error = %v", err)
	}
	defer fb.Dispose()

	fb.Bind()
	driver.Clear(0.0, 1.0, 0.0, 1.0) // Green

	// Test full read
	pixels, err := driver.ReadPixels(fb, image.Rectangle{})
	if err != nil {
		t.Fatalf("ReadPixels() error = %v", err)
	}
	if len(pixels) != 8*8*4 {
		t.Errorf("len(pixels) = %d, want %d", len(pixels), 8*8*4)
	}

	// Test partial read
	pixels, err = driver.ReadPixels(fb, image.Rect(2, 2, 6, 6))
	if err != nil {
		t.Fatalf("ReadPixels() error = %v", err)
	}
	if len(pixels) != 4*4*4 {
		t.Errorf("len(pixels) = %d, want %d", len(pixels), 4*4*4)
	}

	// Test read with clamped region
	pixels, err = driver.ReadPixels(fb, image.Rect(-1, -1, 100, 100))
	if err != nil {
		t.Fatalf("ReadPixels() error = %v", err)
	}
	if len(pixels) != 8*8*4 {
		t.Errorf("len(pixels) = %d, want %d (clamped)", len(pixels), 8*8*4)
	}

	// Test read from nil framebuffer with current set
	pixels, err = driver.ReadPixels(nil, image.Rectangle{})
	if err != nil {
		t.Fatalf("ReadPixels(nil) error = %v", err)
	}
}

func TestMetalDrawCall(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	fb, err := driver.NewFramebuffer(64, 64)
	if err != nil {
		t.Fatalf("NewFramebuffer() error = %v", err)
	}
	defer fb.Dispose()

	fb.Bind()
	driver.Clear(0.0, 0.0, 0.0, 1.0) // Black background

	src, err := LoadShaderSource("solid")
	if err != nil {
		t.Fatalf("LoadShaderSource() error = %v", err)
	}

	shader, err := driver.NewShader(src)
	if err != nil {
		t.Fatalf("NewShader() error = %v", err)
	}
	defer shader.Dispose()

	shader.SetUniform("uTransform", Identity())

	// Create a simple red quad
	// Vertex layout: float2 position, float4 color = 6 floats per vertex
	vertices := []float32{
		// Position (x, y), Color (r, g, b, a)
		-0.5, -0.5, 1.0, 0.0, 0.0, 1.0,
		0.5, -0.5, 1.0, 0.0, 0.0, 1.0,
		0.5, 0.5, 1.0, 0.0, 0.0, 1.0,
		-0.5, 0.5, 1.0, 0.0, 0.0, 1.0,
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}

	// Test draw with various blend modes
	for _, blend := range []BlendMode{BlendNone, BlendAlpha, BlendAdditive} {
		call := DrawCall{
			Shader:   shader,
			Vertices: vertices,
			Indices:  indices,
			Target:   fb,
			Blend:    blend,
		}
		driver.Draw(call)
	}

	// Test draw with scissor
	scissor := image.Rect(10, 10, 54, 54)
	call := DrawCall{
		Shader:   shader,
		Vertices: vertices,
		Indices:  indices,
		Target:   fb,
		Blend:    BlendAlpha,
		Scissor:  &scissor,
	}
	driver.Draw(call)
}

func TestMetalDrawCallWithTexture(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	fb, err := driver.NewFramebuffer(32, 32)
	if err != nil {
		t.Fatalf("NewFramebuffer() error = %v", err)
	}
	defer fb.Dispose()

	fb.Bind()
	driver.Clear(0.0, 0.0, 0.0, 1.0)

	tex, err := driver.NewTexture(4, 4)
	if err != nil {
		t.Fatalf("NewTexture() error = %v", err)
	}
	defer tex.Dispose()

	// Upload white pixels
	pixels := make([]byte, 4*4*4)
	for i := 0; i < len(pixels); i++ {
		pixels[i] = 255
	}
	tex.Upload(pixels, image.Rectangle{})

	src, err := LoadShaderSource("texture")
	if err != nil {
		t.Skipf("texture shader not available: %v", err)
	}

	shader, err := driver.NewShader(src)
	if err != nil {
		t.Fatalf("NewShader() error = %v", err)
	}
	defer shader.Dispose()

	vertices := []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}

	call := DrawCall{
		Shader:   shader,
		Vertices: vertices,
		Indices:  indices,
		Textures: []Texture{tex},
		Target:   fb,
		Blend:    BlendAlpha,
	}
	driver.Draw(call)
}

func TestMetalDrawCallEdgeCases(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	fb, err := driver.NewFramebuffer(32, 32)
	if err != nil {
		t.Fatalf("NewFramebuffer() error = %v", err)
	}
	defer fb.Dispose()

	fb.Bind()

	src, err := LoadShaderSource("solid")
	if err != nil {
		t.Fatalf("LoadShaderSource() error = %v", err)
	}

	shader, err := driver.NewShader(src)
	if err != nil {
		t.Fatalf("NewShader() error = %v", err)
	}
	defer shader.Dispose()

	// Test draw with nil shader
	driver.Draw(DrawCall{
		Shader:   nil,
		Vertices: []float32{0, 0, 1, 1, 1, 1},
		Indices:  []uint16{0},
	})

	// Test draw with empty vertices
	driver.Draw(DrawCall{
		Shader:   shader,
		Vertices: []float32{},
		Indices:  []uint16{0},
	})

	// Test draw with empty indices
	driver.Draw(DrawCall{
		Shader:   shader,
		Vertices: []float32{0, 0, 1, 1, 1, 1},
		Indices:  []uint16{},
	})

	// Test draw with wrong shader type
	driver.Draw(DrawCall{
		Shader:   &stubShader{},
		Vertices: []float32{0, 0, 1, 1, 1, 1},
		Indices:  []uint16{0},
	})
}

func TestMetalNilHandling(t *testing.T) {
	// Test nil driver methods
	var nilDriver *metalDriver
	nilDriver.Clear(1, 0, 0, 1)
	nilDriver.Draw(DrawCall{})
	nilDriver.Dispose()
	nilDriver.run(func() {})

	_, err := nilDriver.NewTexture(1, 1)
	if err != ErrUnsupported {
		t.Errorf("nil driver NewTexture() should return ErrUnsupported")
	}

	_, err = nilDriver.NewFramebuffer(1, 1)
	if err != ErrUnsupported {
		t.Errorf("nil driver NewFramebuffer() should return ErrUnsupported")
	}

	_, err = nilDriver.NewShader(ShaderSource{})
	if err != ErrUnsupported {
		t.Errorf("nil driver NewShader() should return ErrUnsupported")
	}

	_, err = nilDriver.ReadPixels(nil, image.Rectangle{})
	if err != ErrUnsupported {
		t.Errorf("nil driver ReadPixels() should return ErrUnsupported")
	}

	// Test nil texture methods
	var nilTex *metalTexture
	if nilTex.ID() != 0 {
		t.Errorf("nil texture ID() should return 0")
	}
	w, h := nilTex.Size()
	if w != 0 || h != 0 {
		t.Errorf("nil texture Size() should return (0, 0)")
	}
	nilTex.Upload(nil, image.Rectangle{})
	nilTex.Dispose()

	// Test nil framebuffer methods
	var nilFB *metalFramebuffer
	if nilFB.ID() != 0 {
		t.Errorf("nil framebuffer ID() should return 0")
	}
	w, h = nilFB.Size()
	if w != 0 || h != 0 {
		t.Errorf("nil framebuffer Size() should return (0, 0)")
	}
	nilFB.Bind()
	if nilFB.Texture() != nil {
		t.Errorf("nil framebuffer Texture() should return nil")
	}
	nilFB.Dispose()

	// Test nil shader methods
	var nilShader *metalShader
	if nilShader.ID() != 0 {
		t.Errorf("nil shader ID() should return 0")
	}
	nilShader.SetUniform("test", 1.0)
	nilShader.Dispose()
}

func TestMetalDriverDispose(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}

	// Create resources
	tex, _ := driver.NewTexture(4, 4)
	fb, _ := driver.NewFramebuffer(4, 4)

	src, _ := LoadShaderSource("solid")
	shader, _ := driver.NewShader(src)

	// Dispose resources first
	if tex != nil {
		tex.Dispose()
	}
	if fb != nil {
		fb.Dispose()
	}
	if shader != nil {
		shader.Dispose()
	}

	// Dispose driver - should clean up queue and device
	driver.Dispose()

	// Double dispose should be safe
	driver.Dispose()
}

func TestMetalReadTexturePixels(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	// ReadTexturePixels requires the driver to implement this interface
	rtpDriver, ok := driver.(interface {
		ReadTexturePixels(Texture, image.Rectangle) ([]byte, int, int, error)
	})
	if !ok {
		t.Skip("Driver does not implement ReadTexturePixels")
	}

	tex, err := driver.NewTexture(4, 4)
	if err != nil {
		t.Fatalf("NewTexture() error = %v", err)
	}
	defer tex.Dispose()

	// Upload blue pixels
	pixels := make([]byte, 4*4*4)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = 0     // R
		pixels[i+1] = 0   // G
		pixels[i+2] = 255 // B
		pixels[i+3] = 255 // A
	}
	tex.Upload(pixels, image.Rectangle{})

	// Read back
	readPixels, w, h, err := rtpDriver.ReadTexturePixels(tex, image.Rectangle{})
	if err != nil {
		t.Fatalf("ReadTexturePixels() error = %v", err)
	}
	if w != 4 || h != 4 {
		t.Errorf("size = (%d, %d), want (4, 4)", w, h)
	}
	if len(readPixels) != 4*4*4 {
		t.Errorf("len(pixels) = %d, want %d", len(readPixels), 4*4*4)
	}

	// Test partial read
	readPixels, w, h, err = rtpDriver.ReadTexturePixels(tex, image.Rect(1, 1, 3, 3))
	if err != nil {
		t.Fatalf("ReadTexturePixels() partial error = %v", err)
	}
	if len(readPixels) != 2*2*4 {
		t.Errorf("partial len(pixels) = %d, want %d", len(readPixels), 2*2*4)
	}
}

func TestMetalRunMethod(t *testing.T) {
	skipIfNoMetal(t)

	driver, err := newMetalDriver()
	if err != nil {
		t.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	// Test that run executes the function
	executed := false
	md := driver.(*metalDriver)
	md.run(func() {
		executed = true
	})
	if !executed {
		t.Error("run() did not execute the function")
	}

	// Test run with nil function
	md.run(nil)
}

// Benchmark tests

func BenchmarkMetalTextureCreation(b *testing.B) {
	if runtime.GOOS != "darwin" {
		b.Skip("Metal tests only run on macOS")
	}
	if err := objcInit(); err != nil {
		b.Skipf("Metal not available: %v", err)
	}

	driver, err := newMetalDriver()
	if err != nil {
		b.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tex, err := driver.NewTexture(256, 256)
		if err != nil {
			b.Fatalf("NewTexture() error = %v", err)
		}
		tex.Dispose()
	}
}

func BenchmarkMetalClear(b *testing.B) {
	if runtime.GOOS != "darwin" {
		b.Skip("Metal tests only run on macOS")
	}
	if err := objcInit(); err != nil {
		b.Skipf("Metal not available: %v", err)
	}

	driver, err := newMetalDriver()
	if err != nil {
		b.Skipf("Metal driver creation failed: %v", err)
	}
	defer driver.Dispose()

	fb, err := driver.NewFramebuffer(512, 512)
	if err != nil {
		b.Fatalf("NewFramebuffer() error = %v", err)
	}
	defer fb.Dispose()

	fb.Bind()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		driver.Clear(1.0, 0.0, 0.0, 1.0)
	}
}
