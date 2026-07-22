package main

import (
	"context"
	"fmt"
	"os"

	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/gosxscene"
	"m31labs.dev/fluffyui/widgets"
	"m31labs.dev/gosx/scene"
	"m31labs.dev/gosx/scene/preview"
)

func main() {
	view, err := gosxscene.New(demoScene(), gosxscene.Options{
		Preview:  preview.Options{Width: 960, Height: 540},
		Fit:      widgets.ImageFitContain,
		Protocol: widgets.ImageProtocolHalfBlock,
		Label:    "GoSX Scene3D terminal demo",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "render GoSX scene: %v\\n", err)
		os.Exit(1)
	}
	if err := fluffy.RunContext(context.Background(), view); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "run FluffyUI: %v\\n", err)
		os.Exit(1)
	}
}

func demoScene() scene.Props {
	return scene.Props{
		Background: "#0a0d16",
		Camera: scene.PerspectiveCamera{
			Position: scene.Vec3(0, 1.2, 6.2),
			FOV:      40,
			Near:     0.1,
			Far:      40,
		},
		Graph: scene.NewGraph(
			scene.AmbientLight{ID: "ambient", Color: "#b9c8ff", Intensity: 0.3},
			scene.DirectionalLight{ID: "key", Color: "#d8e6ff", Intensity: 1.4, Direction: scene.Vec3(-0.5, -1, -0.6)},
			scene.PointLight{ID: "rim", Color: "#ff6fcf", Intensity: 1.1, Position: scene.Vec3(2.8, 2.3, 2), Range: 12},
			scene.Mesh{
				ID:       "terminal-cube",
				Geometry: scene.BoxGeometry{Width: 2.1, Height: 2.1, Depth: 2.1},
				Material: scene.StandardMaterial{Color: "#7654ff", Roughness: 0.22, Metalness: 0.7, Clearcoat: 0.4},
				Rotation: scene.Rotate(0.32, 0.48, -0.08),
			},
		),
	}
}
