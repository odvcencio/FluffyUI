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
uniform float uBrightness;
uniform float uContrast;
uniform float uSaturation;
uniform float uHue;

vec3 rotateHue(vec3 color, float hue) {
	float cosA = cos(hue);
	float sinA = sin(hue);
	mat3 m = mat3(
		0.213 + cosA*0.787 - sinA*0.213, 0.715 - cosA*0.715 - sinA*0.715, 0.072 - cosA*0.072 + sinA*0.928,
		0.213 - cosA*0.213 + sinA*0.143, 0.715 + cosA*0.285 + sinA*0.140, 0.072 - cosA*0.072 - sinA*0.283,
		0.213 - cosA*0.213 - sinA*0.787, 0.715 - cosA*0.715 + sinA*0.715, 0.072 + cosA*0.928 + sinA*0.072
	);
	return clamp(m * color, 0.0, 1.0);
}

void main() {
	vec4 col = texture(uTexture, vUV);
	if (col.a == 0.0) {
		FragColor = col;
		return;
	}
	vec3 rgb = col.rgb;
	if (uBrightness != 0.0) {
		rgb = clamp(rgb + uBrightness, 0.0, 1.0);
	}
	if (uContrast != 0.0) {
		float factor = 1.0 + uContrast;
		rgb = clamp((rgb - 0.5) * factor + 0.5, 0.0, 1.0);
	}
	if (uSaturation != 0.0) {
		float luma = dot(rgb, vec3(0.2126, 0.7152, 0.0722));
		rgb = clamp(vec3(luma) + (rgb - vec3(luma)) * (1.0 + uSaturation), 0.0, 1.0);
	}
	if (uHue != 0.0) {
		rgb = rotateHue(rgb, uHue);
	}
	FragColor = vec4(rgb, col.a);
}
