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
uniform float uRadius;
uniform float uSoftness;
void main() {
	vec4 col = texture(uTexture, vUV);
	vec2 center = vec2(0.5, 0.5);
	float dist = distance(vUV, center);
	float edge = smoothstep(uRadius, uRadius + uSoftness, dist);
	col.rgb *= (1.0 - edge);
	FragColor = col;
}
