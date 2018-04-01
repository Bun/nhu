function _setInterval(f, t) {
    setInterval(f, t);
    f();
}

function sourceFetcher(path, cb) {
    var req, rev, start, source;
    function setup() {
        start = performance.now();
        req = new XMLHttpRequest();
        req.addEventListener('load', load);
        req.open('GET', path + (rev ? ('?' + rev) : ''));
        req.send();
    }

    function load() {
        if (req.readyState !== 4)
            return
        try {
            const status = req.status;
            if (status !== 200) {
                console.error('status', status);
                return;
            }
            // Updated?
            if (source === this.responseText)
                return;
            source = this.responseText;
            rev = req.getResponseHeader('X-Revision');
            cb(source, rev);
        } finally {
            next();
        }
    }

    function next() {
        const now = performance.now();
        if ((now - start) < 1000) {
            console.log('delay next');
            setTimeout(setup, 1000);
        } else {
            console.log('imm next');
            setup();
        }
    }

    setup();
}

document.addEventListener('DOMContentLoaded', function() {
    const fhead = `precision mediump float;
uniform vec2 u_resolution;
uniform float u_time;
`;
    const vsrc = `attribute vec4 position;
void main() {
    gl_Position = position;
}`;
    const fsrc = `void main() { gl_FragColor = vec4(0., 0.2, 0., 1.); }`;

    // TODO: autoscale
    const width = 720;
    const height = 500;

    const d = document.createElement('canvas');
    d.setAttribute('width', width);
    d.setAttribute('height', height);
    const gl = d.getContext('webgl');
    if (!gl) {
        alert("No WebGL");
        return;
    }

    /** OGL SETUP **/

    const prog = gl.createProgram();
    const vars = {
        vertex: compileShader(gl.VERTEX_SHADER, vsrc),
        fragment: compileShader(gl.FRAGMENT_SHADER, fsrc)
    };

    gl.attachShader(prog, vars.vertex);
    gl.attachShader(prog, vars.fragment);
    link();
    gl.useProgram(prog);

    function link() {
        gl.linkProgram(prog);
        vars.position = gl.getAttribLocation(prog, 'position');
        vars.u_resolution = gl.getUniformLocation(prog, 'u_resolution');
        vars.u_time = gl.getUniformLocation(prog, 'u_time');
    }

    /** SHADER **/

    sourceFetcher('shader/fragment',
        function(text, revision) {
            console.log('Update shader', revision);
            const z = compileShader(gl.FRAGMENT_SHADER, text);
            if (z !== null) {
                gl.detachShader(prog, vars.fragment);
                gl.deleteShader(vars.fragment);
                gl.attachShader(prog, z);
                link();
                vars.fragment = z;
            }
        });

    /** CANVAS **/

    const buff = gl.createBuffer();
    const positions = [-1, -1, 1, -1, 1, 1, -1, 1];
    gl.bindBuffer(gl.ARRAY_BUFFER, buff);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(positions), gl.STATIC_DRAW);
    gl.vertexAttribPointer(vars.position, 2, gl.FLOAT, false, 0, 0);
    gl.enableVertexAttribArray(vars.position);

    /** DRAW **/

    var start = performance.now() / 1000;
    requestAnimationFrame(draw);

    function draw() {
        const u_time = (performance.now() / 1000) - start;
        gl.useProgram(prog);
        gl.uniform1f(vars.u_time, u_time);
        gl.uniform2f(vars.u_resolution, width, height);
        gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
        gl.drawArrays(gl.TRIANGLE_FAN, 0, 4);
        requestAnimationFrame(draw);
    }

    function compileShader(type, source) {
        if (type === gl.FRAGMENT_SHADER) {
            source = fhead + '\n' + source;
        }
        const shader = gl.createShader(type);
        gl.shaderSource(shader, source);
        gl.compileShader(shader);
        if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
            console.error('An error occurred compiling the shaders: ' + gl.getShaderInfoLog(shader));
            gl.deleteShader(shader);
            return null;
        }
        return shader;
    }

    document.body.appendChild(d);
});
