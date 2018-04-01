package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	width  = 720
	height = 500
	fps    = 30
)

func init() {
	flag.IntVar(&fps, "fps", fps, "Frames per second")
}

type GlView struct {
	source chan string
}

func renderThread(source chan string) {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL:", err)
	}

	win := window()

	prog := gl.CreateProgram()
	if vshader, err := compileShader(defaultVertexShader, gl.VERTEX_SHADER); err != nil {
		log.Fatalln("Failed to compile vertex shader:", err)
	} else {
		gl.AttachShader(prog, vshader)
	}
	frag := <-source
	fshader, err := compileShader(frag, gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatalln("Failed to compile fragment shader:", err)
	} else {
		gl.AttachShader(prog, fshader)
	}
	gl.LinkProgram(prog)
	gl.UseProgram(prog)
	uTime := gl.GetUniformLocation(prog, gl.Str("u_time\x00"))
	uResolution := gl.GetUniformLocation(prog, gl.Str("u_resolution\x00"))

	start := time.Now()
	var t time.Time

	cvs := makeCanvas()
	gl.BindVertexArray(cvs)

	for !win.ShouldClose() {
		select {
		case frag := <-source:
			if z, err := compileShader(frag, gl.FRAGMENT_SHADER); err != nil {
				log.Println("Failed to update fragment shader:", err)
			} else {
				gl.DetachShader(prog, fshader)
				gl.DeleteShader(fshader)
				gl.AttachShader(prog, z)
				fshader = z
				gl.LinkProgram(prog)
				uTime = gl.GetUniformLocation(prog, gl.Str("u_time\x00"))
				uResolution = gl.GetUniformLocation(prog, gl.Str("u_resolution\x00"))
				log.Println("Shader updated")
			}
		default:
		}

		t = time.Now()

		ftime := float32(t.Sub(start)) / float32(time.Second)
		gl.Uniform1f(uTime, ftime)
		gl.Uniform2f(uResolution, float32(width), float32(height))

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

		glfw.PollEvents()
		win.SwapBuffers()

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}

	// TODO: terminate program in a nicer way
	os.Exit(0)
}

func NewGlView() *GlView {
	source := make(chan string, 1)
	go renderThread(source)
	return &GlView{source}
}

func (g *GlView) Update(s string) {
	g.source <- s
}

func window() *glfw.Window {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(width, height, "NHU", nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	window.MakeContextCurrent()
	return window
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	if shaderType == gl.FRAGMENT_SHADER {
		source = `#version 400
uniform vec2 u_resolution;
uniform float u_time;
` + source

	}
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("Failed to compile: %v", log)
	}
	return shader, nil
}

func makeCanvas() uint32 {
	points := []float32{-1, -1, 1, -1, 1, 1, -1, 1}
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)

	return vao
}
