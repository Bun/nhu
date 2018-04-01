package main

// Default shaders
const (
	defaultVertexShader = `#version 400
in vec3 position;
void main() {
	gl_Position = vec4(position, 1.0);
}
`

	defaultFragmentShader = `// NHU

void main() {
    vec2 uv = (gl_FragCoord.xy / u_resolution) - 0.5;
    uv.x *= u_resolution.x / u_resolution.y;

    gl_FragColor = vec4(fract(u_time), sin(uv.x), 0.0, 1.0);
}
`
)
