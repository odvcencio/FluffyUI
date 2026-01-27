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
uniform vec2 uOffset;
void main() {
	vec2 off = uOffset;
	vec4 base = texture(uTexture, vUV);
	vec4 r = texture(uTexture, vUV - off);
	vec4 b = texture(uTexture, vUV + off);
	FragColor = vec4(r.r, base.g, b.b, base.a);
}
