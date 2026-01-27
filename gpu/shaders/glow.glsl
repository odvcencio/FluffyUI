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
uniform vec4 uGlowColor;
void main() {
	vec4 base = texture(uTexture, vUV);
	FragColor = vec4(uGlowColor.rgb, base.a * uGlowColor.a);
}
