package gpu

import (
	"embed"
	"errors"
	"path"
	"strings"
)

//go:embed shaders/*.glsl shaders/*.metal
var shaderFS embed.FS

// LoadShaderSource loads embedded shader source by name (e.g. "solid" or "solid.glsl").
func LoadShaderSource(name string) (ShaderSource, error) {
	base := normalizeShaderName(name)
	if base == "" {
		return ShaderSource{}, errors.New("shader name required")
	}
	src := ShaderSource{}
	if glslBytes, err := shaderFS.ReadFile("shaders/" + base + ".glsl"); err == nil {
		vertex, fragment, splitErr := splitGLSL(string(glslBytes))
		if splitErr != nil {
			return ShaderSource{}, splitErr
		}
		src.GLSL = GLSLSource{Vertex: vertex, Fragment: fragment}
	}
	if metalBytes, err := shaderFS.ReadFile("shaders/" + base + ".metal"); err == nil {
		src.Metal = string(metalBytes)
	}
	if src.GLSL.Vertex == "" && src.Metal == "" {
		return ShaderSource{}, errors.New("shader not found")
	}
	return src, nil
}

func normalizeShaderName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.TrimPrefix(name, "shaders/")
	ext := path.Ext(name)
	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}

func splitGLSL(src string) (string, string, error) {
	var vertex strings.Builder
	var fragment strings.Builder
	mode := ""
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			if strings.Contains(trimmed, "Vertex shader") {
				mode = "vertex"
				continue
			}
			if strings.Contains(trimmed, "Fragment shader") {
				mode = "fragment"
				continue
			}
		}
		switch mode {
		case "vertex":
			vertex.WriteString(line)
			vertex.WriteString("\n")
		case "fragment":
			fragment.WriteString(line)
			fragment.WriteString("\n")
		}
	}
	if vertex.Len() == 0 || fragment.Len() == 0 {
		return "", "", errors.New("glsl shader missing vertex or fragment section")
	}
	return strings.TrimSpace(vertex.String()) + "\n", strings.TrimSpace(fragment.String()) + "\n", nil
}
