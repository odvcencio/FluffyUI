// Vertex shader
#version 330 core
layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aUV;
out vec2 vUV;
uniform mat3 uTransform;
void main() {
	vec3 pos = uTransform * vec3(aPos, 1.0);
	gl_Position = vec4(pos.xy, 0.0, 1.0);
	vUV = aUV;
}

// Fragment shader
#version 330 core
in vec2 vUV;
out vec4 FragColor;
uniform sampler2D uTexture;
void main() {
	FragColor = texture(uTexture, vUV);
}
