#include <metal_stdlib>
using namespace metal;

struct VertexIn {
	float2 position [[attribute(0)]];
	float2 uv [[attribute(1)]];
};

struct VertexOut {
	float4 position [[position]];
	float2 uv;
};

vertex VertexOut vertex_main(VertexIn in [[stage_in]]) {
	VertexOut out;
	out.position = float4(in.position, 0.0, 1.0);
	out.uv = in.uv;
	return out;
}

float3 rotateHue(float3 color, float hue) {
	float cosA = cos(hue);
	float sinA = sin(hue);
	float3x3 m = float3x3(
		float3(0.213 + cosA * 0.787 - sinA * 0.213, 0.213 - cosA * 0.213 + sinA * 0.143, 0.213 - cosA * 0.213 - sinA * 0.787),
		float3(0.715 - cosA * 0.715 - sinA * 0.715, 0.715 + cosA * 0.285 + sinA * 0.140, 0.715 - cosA * 0.715 + sinA * 0.715),
		float3(0.072 - cosA * 0.072 + sinA * 0.928, 0.072 - cosA * 0.072 - sinA * 0.283, 0.072 + cosA * 0.928 + sinA * 0.072)
	);
	return clamp(m * color, 0.0, 1.0);
}

fragment float4 fragment_main(VertexOut in [[stage_in]],
							  texture2d<float> tex [[texture(0)]],
							  constant float& uBrightness [[buffer(1)]],
							  constant float& uContrast [[buffer(2)]],
							  constant float& uSaturation [[buffer(3)]],
							  constant float& uHue [[buffer(4)]]) {
	constexpr sampler s(address::clamp_to_edge, filter::linear);
	float4 col = tex.sample(s, in.uv);
	if (col.a == 0.0) {
		return col;
	}
	float3 rgb = col.rgb;
	if (uBrightness != 0.0) {
		rgb = clamp(rgb + uBrightness, 0.0, 1.0);
	}
	if (uContrast != 0.0) {
		float factor = 1.0 + uContrast;
		rgb = clamp((rgb - 0.5) * factor + 0.5, 0.0, 1.0);
	}
	if (uSaturation != 0.0) {
		float luma = dot(rgb, float3(0.2126, 0.7152, 0.0722));
		rgb = clamp(float3(luma) + (rgb - float3(luma)) * (1.0 + uSaturation), 0.0, 1.0);
	}
	if (uHue != 0.0) {
		rgb = rotateHue(rgb, uHue);
	}
	return float4(rgb, col.a);
}
