#include <metal_stdlib>
using namespace metal;

struct VertexIn {
	float2 position [[attribute(0)]];
	float4 color [[attribute(1)]];
};

struct VertexOut {
	float4 position [[position]];
	float4 color;
};

vertex VertexOut vertex_main(VertexIn in [[stage_in]], constant float4x4& uTransform [[buffer(1)]]) {
	VertexOut out;
	float4 pos = uTransform * float4(in.position, 0.0, 1.0);
	out.position = pos;
	out.color = in.color;
	return out;
}

fragment float4 fragment_main(VertexOut in [[stage_in]]) {
	return in.color;
}
