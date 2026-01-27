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

fragment float4 fragment_main(VertexOut in [[stage_in]],
							  texture2d<float> tex [[texture(0)]],
							  constant float& uPixelSize [[buffer(1)]],
							  constant float2& uResolution [[buffer(2)]]) {
	constexpr sampler s(address::clamp_to_edge, filter::linear);
	float2 size = max(float2(uPixelSize), float2(1.0)) / uResolution;
	float2 uv = floor(in.uv / size) * size + size * 0.5;
	return tex.sample(s, uv);
}
