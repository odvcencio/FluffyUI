package gosxscene

import (
	"math"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/widgets"
	"m31labs.dev/gosx/scene"
	"m31labs.dev/gosx/scene/preview"
)

func TestNew_RendersTypedSceneIntoWidget(t *testing.T) {
	view, err := New(testScene(), Options{
		Preview:  previewOptions(96, 64),
		Fit:      widgets.ImageFitContain,
		Protocol: widgets.ImageProtocolHalfBlock,
		Label:    "GoSX terminal scene",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if view.Result() == nil || view.Result().Image == nil {
		t.Fatal("New() did not retain a rendered GoSX frame")
	}
	if got := view.Result().Image.Bounds().Dx(); got != 96 {
		t.Fatalf("frame width = %d, want 96", got)
	}
	if got := view.Result().Image.Bounds().Dy(); got != 64 {
		t.Fatalf("frame height = %d, want 64", got)
	}
	if got := len(view.Result().Bundle.InstancedMeshes); got != 1 {
		t.Fatalf("render bundle mesh count = %d, want 1", got)
	}
	if view.Label() != "GoSX terminal scene" {
		t.Fatalf("label = %q", view.Label())
	}
	var widget runtime.Widget = view
	if widget == nil {
		t.Fatal("Scene does not satisfy runtime.Widget")
	}
}

func TestScene_RenderFrameUpdatesResult(t *testing.T) {
	view, err := New(testScene(), Options{Preview: previewOptions(80, 48)})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	before := view.Result()
	if err := view.RenderFrame(1.25); err != nil {
		t.Fatalf("RenderFrame() error = %v", err)
	}
	if view.Result() == before {
		t.Fatal("RenderFrame() did not replace the renderer result")
	}
	if got, want := view.Result().Bundle.Camera.FOV, 42*math.Pi/180; math.Abs(got-want) > 1e-9 {
		t.Fatalf("camera FOV = %v, want %v", got, want)
	}
}

func testScene() scene.Props {
	return scene.Props{
		Background: "#10131d",
		Camera: scene.PerspectiveCamera{
			Position: scene.Vec3(0, 0.5, 5),
			FOV:      42,
			Near:     0.1,
			Far:      20,
		},
		Graph: scene.NewGraph(
			scene.AmbientLight{ID: "ambient", Color: "#ffffff", Intensity: 0.4},
			scene.DirectionalLight{ID: "key", Color: "#9fc5ff", Intensity: 1.2, Direction: scene.Vec3(-0.4, -1, -0.7)},
			scene.Mesh{
				ID:       "cube",
				Geometry: scene.BoxGeometry{Width: 1.8, Height: 1.8, Depth: 1.8},
				Material: scene.StandardMaterial{Color: "#7c5cff", Roughness: 0.26, Metalness: 0.58},
				Rotation: scene.Rotate(0.3, 0.5, 0),
			},
		),
	}
}

func previewOptions(width, height int) preview.Options {
	return preview.Options{Width: width, Height: height, DisableShadows: true}
}
