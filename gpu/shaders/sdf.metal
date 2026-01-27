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

vertex VertexOut vertex_main(VertexIn in [[stage_in]], constant float4x4& uTransform [[buffer(1)]]) {
	VertexOut out;
	float4 pos = uTransform * float4(in.position, 0.0, 1.0);
	out.position = pos;
	out.uv = in.uv;
	return out;
}

fragment float4 fragment_main(VertexOut in [[stage_in]], texture2d<float> tex [[texture(0)]], constant float4& uColor [[buffer(2)]], constant float& uThreshold [[buffer(3)]]) {
	constexpr sampler s(address::clamp_to_edge, filter::linear);
	float dist = tex.sample(s, in.uv).r;
	float alpha = smoothstep(uThreshold - 0.02, uThreshold + 0.02, dist);
	return float4(uColor.rgb, uColor.a * alpha);
}
