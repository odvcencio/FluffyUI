// Vertex shader
#version 330 core
layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aUV;
out vec2 vUV;
void main() {
	gl_Position = vec4(aPos, 0.0, 1.0);
	vUV = aUV;
}

// Fragment shader
#version 330 core
in vec2 vUV;
out vec4 FragColor;
uniform sampler2D uTexture;
uniform float uPixelSize;
uniform vec2 uResolution;
void main() {
	vec2 size = max(vec2(uPixelSize), vec2(1.0)) / uResolution;
	vec2 uv = floor(vUV / size) * size + size * 0.5;
	FragColor = texture(uTexture, uv);
}
