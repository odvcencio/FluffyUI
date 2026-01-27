// Vertex shader
#version 330 core
layout(location = 0) in vec2 aPos;
layout(location = 1) in vec4 aColor;
out vec4 vColor;
uniform mat3 uTransform;
void main() {
	vec3 pos = uTransform * vec3(aPos, 1.0);
	gl_Position = vec4(pos.xy, 0.0, 1.0);
	vColor = aColor;
}

// Fragment shader
#version 330 core
in vec4 vColor;
out vec4 FragColor;
void main() {
	FragColor = vColor;
}
