package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/go-gl/gl/v3.2-core/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	// options controlling size of earth
	radius = 1.0       // radius of earth model
	a      = 6378137.0 // earth radius in m
	//rf     = 298.257223563 // wgs-84 flattening factor (not used)

	// texture paths for earth model
	diffusePath  = "../../assets/textures/earth2.jpg"
	specularPath = "../../assets/textures/spec_map.png"
	borderPath   = "../../assets/textures/night_lights.png"
	cloudPath    = "../../assets/textures/clouds.jpg"

	// shader paths for earth, objects, clouds
	vertexShaderPath         = "../../assets/shaders/vertexshader.glslv"
	fragmentShaderPath       = "../../assets/shaders/fragmentshader.glslf"
	objectVertexShaderPath   = "../../assets/shaders/objectvertexshader.glslv"
	objectFragmentShaderPath = "../../assets/shaders/objectfragmentshader.glslf"
	cloudVertexShaderPath    = "../../assets/shaders/cloudvertexshader.glslv"
	cloudFragmentShaderPath  = "../../assets/shaders/cloudfragmentshader.glslf"

	cGreen = "\x1B[32m"
	cNorm  = "\x1B[0m"
)

//Camera stores position, front/up vectors, fov of camera
type Camera struct {
	Pos   mgl32.Vec3
	Front mgl32.Vec3
	Up    mgl32.Vec3
	fov   int
}

//Mouse stores lastx, lasty, other mouse variables
type Mouse struct {
	lastX      float32
	lastY      float32
	firstMouse bool
	yaw        float32
	pitch      float32
}

//GameState stores the state of the application, whether it is receiving input, reloading vertices
type GameState struct {
	resetting bool
	inputting bool

	showEarth          bool
	showLines          bool
	showPoints         bool
	showOrbits         bool
	enableAntialiasing bool
	enableBlending     bool
	fps                int
}

var (
	// default globe resolution (modelres=1)
	stackCount       = 16
	sectorCount      = 32
	cloudStackCount  = 16
	cloudSectorCount = 32

	// screen width and height
	width  = 800
	height = 600

	// used to store mouse information for FPS camera
	mouse Mouse

	// default point size
	pointSize = 8.0

	// default colors
	lightColor      = mgl32.Vec3{1.0, 1.0, 0.8} // "sun" color
	objectColor     = mgl32.Vec3{1.0, 1.0, 1.0} // earth color tint
	backgroundColor = mgl32.Vec3{5 / 255, 20 / 255, 50 / 255}
	atmoColor       = mgl32.Vec3{94.0 / 255.0, 149.0 / 255.0, 239.0 / 255.0} // atmosphere color
	atmoColor2      = mgl32.Vec3{55.0 / 255.0, 87.0 / 255.0, 157.0 / 255.0}  // atmosphere color

	// camera object
	camera Camera

	//position of "sun" light, ambient light strength
	lightPos        = mgl32.Vec3{200.0, 50.0, 200.0}
	ambientStrength = 0.2

	//default movement speed
	moveSpeed float32 = 3.0

	// angle for earth rotation
	angleZ float32

	// default values to show earth, lines, points, orbits, antialiasing, blending, resetting, inputting, fps
	state GameState

	// default file to read
	visualOutputPath = "../../examples/diorama-visual-output.kml"

	// default MSAA samples, alpha blending
	samples = 8
	alpha   = 0.6
)

func init() {
	//required
	runtime.LockOSThread()

	camera.Pos = mgl32.Vec3{0.0, 0.0, 15.0}
	camera.Front = mgl32.Vec3{0.0, 0.0, -1.0}
	camera.Up = mgl32.Vec3{0.0, 1.0, 0.0}

	camera.fov = 35

	mouse.firstMouse = true
	mouse.yaw = -90.0

	state.showEarth = true
	state.showPoints = true
	state.showLines = true
	state.showOrbits = true

	state.resetting = true
	state.inputting = false

	state.enableBlending = true
	state.enableAntialiasing = true
}

func main() {
	// define all flag values
	filePath := flag.String("file", visualOutputPath, "/path/to/kml")
	widthF := flag.Int("width", width, "width")
	heightF := flag.Int("height", height, "height")
	fovF := flag.Int("fov", camera.fov, "field of view")
	resF := flag.Int("modelres", 4, "model resolution multiplier")
	psF := flag.Float64("ps", pointSize, "point size")
	samplesF := flag.Int("samples", samples, "number of MSAA samples")
	ambientF := flag.Float64("ambient", ambientStrength, "strength of ambient lighting")
	alphaF := flag.Float64("alpha", alpha, "line transparency")
	gridF := flag.Bool("grid", false, "generate grid")

	// parse flags
	fmt.Println("Parsing flags...")
	flag.Parse()
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	//use flag values to set required variables
	visualOutputPath = *filePath
	width = *widthF
	height = *heightF
	camera.fov = *fovF
	res := *resF
	stackCount = stackCount * res
	sectorCount = sectorCount * res
	cloudStackCount = cloudStackCount * res
	cloudSectorCount = cloudSectorCount * res

	ambientStrength = *ambientF
	alpha = *alphaF

	pointSize = *psF
	samples = *samplesF

	grid := *gridF

	// initiate glfw and OpenGL
	fmt.Println("Initializing GLFW...")
	win := initGlfw()
	defer glfw.Terminate()
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)
	fmt.Println("Initializing OpenGL...")
	initOpenGL()
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// create the shader programs for each class of objects
	fmt.Println("Generating shader programs...")
	globeProgram := newProgram(vertexShaderPath, fragmentShaderPath)
	objectProgram := newProgram(objectVertexShaderPath, objectFragmentShaderPath)
	cloudProgram := newProgram(cloudVertexShaderPath, cloudFragmentShaderPath)
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// define vertices for the default axis
	axis := []float32{
		// positions         // colors
		0.0, 0.0, 0.0, 1.0, 0.0, 1.0,
		20.0, 0.0, 0.0, 1.0, 0.0, 1.0, // purple x
		0.0, 0.0, 0.0, 1.0, 0.0, 1.0,
		-20.0, 0.0, 0.0, 1.0, 0.0, 1.0,

		0.0, 0.0, 0.0, 0.0, 1.0, 1.0,
		0.0, 0.0, 20.0, 0.0, 1.0, 1.0, // cyan z
		0.0, 0.0, 0.0, 0.0, 1.0, 1.0,
		0.0, 0.0, -20.0, 0.0, 1.0, 1.0,

		0.0, 0.0, 0.0, 0.2, 0.8, 0.0,
		0.0, 20.0, 0.0, 0.2, 0.8, 0.0, // green y
		0.0, 0.0, 0.0, 0.2, 0.8, 0.0,
		0.0, -20.0, 0.0, 0.2, 0.8, 0.0,
	}

	if grid {
		axis = append(axis, genGrid(30, 30, -5, 5, 5, -5)...)
	}

	lineStart := len(axis) / 6

	// init main vertex array
	objectVertices := axis

	// define variables to store locations of different types of data in the vertex array
	pointStart := len(objectVertices)
	orbitStart := len(objectVertices)

	// generate two spheres, one for the globe and one for the clouds
	fmt.Println("Generating sphere vertices...")
	earthVertices, earthIndices := generateSphere(sectorCount, stackCount, radius)
	cloudVertices, cloudIndices := generateSphere(cloudSectorCount, cloudStackCount, radius+0.003)
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	fmt.Println("Generating vertex array objects...")
	// generate a vertex array object to store the object data
	lineVertexArray := makeVaoColoredLines(objectVertices, nil, 6*4)

	// generate two vertex array objects for the earth and the clouds
	earthVertexArray := makeVaoEarth(earthVertices, earthIndices, 8*4)
	cloudVertexArray := makeVaoEarth(cloudVertices, cloudIndices, 8*4)
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// generate necessary textures for the earth and clouds
	fmt.Println("Generating textures...")
	texture := generateTexture(diffusePath)
	texture2 := generateTexture(borderPath)
	specMap := generateTexture(specularPath)
	cloudTexture := generateTexture(cloudPath)
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// ################## SETUP GLOBE UNIFORMS ##################
	fmt.Println("Setting up globe uniform variables...")
	gl.UseProgram(globeProgram)

	// setup scene projection
	projection := mgl32.Perspective(mgl32.DegToRad(float32(camera.fov)), float32(width)/float32(height), 0.01, 100.0)
	_ = setUniform(globeProgram, projection, "projection")

	// setup camera
	cameraMat := mgl32.LookAtV(camera.Pos, camera.Front, camera.Up)
	globeCameraUniform := setUniform(globeProgram, cameraMat, "camera")

	// rotate earth to correct orientation
	model := mgl32.HomogRotate3D(mgl32.DegToRad(-90.0), mgl32.Vec3{1, 0, 0})
	globeModelUniform := setUniform(globeProgram, model, "model")

	// define color of earth, ambient light strength, atmosphere color, light color, light position
	_ = setUniform(globeProgram, objectColor, "objectColor")
	_ = setUniform(globeProgram, ambientStrength, "ambientStrength")
	_ = setUniform(globeProgram, atmoColor, "atmoColor")
	_ = setUniform(globeProgram, atmoColor2, "atmoColor2")
	_ = setUniform(globeProgram, lightColor, "lightColor")
	_ = setUniform(globeProgram, lightPos, "lightPos")

	// set earth textures
	gl.Uniform1i(gl.GetUniformLocation(globeProgram, gl.Str("ourTexture2\x00")), 0)
	gl.Uniform1i(gl.GetUniformLocation(globeProgram, gl.Str("specMap\x00")), 2)
	gl.Uniform1i(gl.GetUniformLocation(globeProgram, gl.Str("ourTexture\x00")), 1)
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// view position
	globeViewPosUniform := setUniform(globeProgram, camera.Pos, "viewPos")

	// ################## SETUP OBJECT UNIFORMS ##################
	fmt.Println("Setting up object uniform variables...")
	gl.UseProgram(objectProgram)

	// rotate objects same as earth
	modelo := mgl32.HomogRotate3D(mgl32.DegToRad(float32(90)), mgl32.Vec3{1, 0, 0}).
		Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(float32(180)), mgl32.Vec3{0, 1, 0})).
		Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(float32(angleZ)), mgl32.Vec3{0, 0, 1}))
	objectModelUniform := setUniform(objectProgram, modelo, "model")

	// set projection, alpha
	_ = setUniform(objectProgram, projection, "projection")
	_ = setUniform(objectProgram, alpha, "alpha")

	// view position
	objectCameraUniform := setUniform(objectProgram, cameraMat, "camera")
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// ################## SETUP CLOUD UNIFORMS ##################
	fmt.Println("Setting up cloud uniform variables...")
	gl.UseProgram(cloudProgram)

	//set cloud texture
	gl.Uniform1i(gl.GetUniformLocation(cloudProgram, gl.Str("cloudTexture\x00")), 3)

	// rotate clouds to correct orientation
	modelc := mgl32.HomogRotate3D(mgl32.DegToRad(-90.0), mgl32.Vec3{1, 0, 0})
	cloudModelUniform := setUniform(cloudProgram, modelc, "model")

	// set projection, ambient, lightpos
	_ = setUniform(cloudProgram, projection, "projection")
	_ = setUniform(cloudProgram, ambientStrength, "ambientStrength")
	_ = setUniform(cloudProgram, lightPos, "lightPos")

	//setup camera, view position
	cloudCameraUniform := setUniform(cloudProgram, cameraMat, "camera")
	cloudViewPosUniform := setUniform(cloudProgram, camera.Pos, "viewPos")
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	//################## SETUP GLOBAL OPENGL OPTIONS ##################
	fmt.Println("Setting up global OpenGL config...")
	// set point size
	gl.PointSize(float32(pointSize))

	// background color
	gl.ClearColor(backgroundColor[0], backgroundColor[1], backgroundColor[2], 1.0)

	// disable culling
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Disable(gl.CULL_FACE)

	//enable smooth line, MSAA
	gl.Enable(gl.LINE_SMOOTH)
	gl.LineWidth(1)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	//needed for consistant camera speed
	deltaTime := 0.0
	lastFrame := 0.0

	lastTime := glfw.GetTime()
	nbFrames := 0

	fmt.Println("Starting GUI on separate thread...")
	//start the gui in a separate goroutine
	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		//gui(win)
	}()
	fmt.Printf("%s[DONE]%s\n", cGreen, cNorm)

	// enter the main render loop
	fmt.Println("Entering main render loop...")
	for !win.ShouldClose() {
		// trigger if vertex data needs to be updated (changed in GUI)
		if state.resetting {
			objectVertices = axis
			kmlVertices := []float32{}

			kmlVertices, pointStart, orbitStart = interpretSelected()
			objectVertices = append(objectVertices, kmlVertices...)

			// generate vertex array for line/point object
			lineVertexArray = makeVaoColoredLines(objectVertices, nil, 6*4)

			state.resetting = false
		}
		// enable/disable antialiasing
		if state.enableAntialiasing {
			gl.Enable(gl.MULTISAMPLE)
		} else {
			gl.Disable(gl.MULTISAMPLE)
		}
		//clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		glfw.PollEvents()

		// used for constant movement speed
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		// frame counter for FPS counter
		nbFrames++
		if currentFrame-lastTime >= 1.0 { // If last prinf() was more than 1 sec ago
			// printf and reset timer
			state.fps = nbFrames
			nbFrames = 0
			lastTime += 1.0
		}

		processInput(win, deltaTime)
		//update matrices
		cameraMat = mgl32.LookAtV(camera.Pos, camera.Pos.Add(camera.Front), camera.Up)
		model = mgl32.HomogRotate3D(mgl32.DegToRad(-90.0), mgl32.Vec3{1, 0, 0}).
			Mul4(mgl32.HomogRotate3D(float32(mgl32.DegToRad(float32(angleZ))), mgl32.Vec3{0, 0, 1}))

		//render globe
		if state.showEarth {
			gl.UseProgram(globeProgram)
			gl.UniformMatrix4fv(globeModelUniform, 1, false, &model[0])
			gl.UniformMatrix4fv(globeCameraUniform, 1, false, &cameraMat[0])
			gl.Uniform3fv(globeViewPosUniform, 1, &camera.Pos[0])
			//gl.Uniform3fv(lightPosUniform, 1, &lightPos[0])

			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, texture2)

			gl.ActiveTexture(gl.TEXTURE1)
			gl.BindTexture(gl.TEXTURE_2D, texture)

			gl.ActiveTexture(gl.TEXTURE2)
			gl.BindTexture(gl.TEXTURE_2D, specMap)

			gl.BindVertexArray(earthVertexArray)

			gl.DrawElements(gl.TRIANGLES, int32((stackCount*stackCount-2)*sectorCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
		}

		//render objects
		if state.showLines || state.showPoints || state.showOrbits {
			gl.UseProgram(objectProgram)
			gl.UniformMatrix4fv(objectCameraUniform, 1, false, &cameraMat[0])
			gl.UniformMatrix4fv(objectModelUniform, 1, false, &model[0])
			gl.BindVertexArray(lineVertexArray)
			// //gl.DrawArrays(gl.LINES, 0, int32(len(vertices)))
			if state.showLines {
				gl.DrawArrays(gl.LINES, 0, int32(pointStart/6+lineStart))
			}

			if state.showPoints {
				gl.DrawArrays(gl.POINTS, int32(pointStart/6+lineStart), int32(orbitStart/6-(pointStart/6)))
				//fmt.Println(basisStart, pointStart)
			}

			//gl.BindVertexArray(orbitVertexArray)
			if state.showOrbits {
				// 	//gl.DrawElements(gl.LINES, int32(len(orbitVertices)/6), gl.UNSIGNED_INT, gl.PtrOffset(0))
				gl.DrawArrays(gl.LINES, int32(orbitStart/6+lineStart), int32((len(objectVertices)-orbitStart)/6))
			}
		}

		//render clouds
		gl.BindVertexArray(cloudVertexArray)
		if state.enableBlending && state.showEarth {
			gl.Enable(gl.BLEND)

			gl.UseProgram(cloudProgram)
			gl.UniformMatrix4fv(cloudCameraUniform, 1, false, &cameraMat[0])
			gl.UniformMatrix4fv(cloudModelUniform, 1, false, &model[0])
			gl.Uniform3fv(cloudViewPosUniform, 1, &camera.Pos[0])

			gl.ActiveTexture(gl.TEXTURE3)
			gl.BindTexture(gl.TEXTURE_2D, cloudTexture)

			gl.DrawElements(gl.TRIANGLES, int32((cloudStackCount*cloudStackCount-2)*cloudSectorCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
		} else if state.enableBlending {
			gl.Enable(gl.BLEND)
		} else {
			gl.Disable(gl.BLEND)
		}

		//collision detection for earth
		d := math.Sqrt(math.Pow(float64(camera.Pos[0]), 2) + math.Pow(float64(camera.Pos[1]), 2) + math.Pow(float64(camera.Pos[2]), 2))
		if d < radius+(float64(moveSpeed)/111+0.02) {
			camera.Pos = camera.Pos.Add(camera.Pos.Mul(radius + ((moveSpeed)/111 + 0.02) - float32(d)))
		}

		win.SwapBuffers()
	}
}

// sets a uniform value in the given shader program. supports mat4, vec3, float
func setUniform(program uint32, value interface{}, id string) int32 {
	uniform := gl.GetUniformLocation(program, gl.Str(id+"\x00"))
	switch v := value.(type) {
	case mgl32.Mat4:
		gl.UniformMatrix4fv(uniform, 1, false, &v[0])
	case mgl32.Vec3:
		gl.Uniform3fv(uniform, 1, &v[0])
	case float64:
		gl.Uniform1f(uniform, float32(v))
	}
	return uniform
}

// initGlfw initializes glfw and returns a Window to use
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	//glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.Samples, samples)

	window, err := glfw.CreateWindow(width, height, "rkmlviewer", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	//window.SetFramebufferSizeCallback(fbCallback)
	//window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetFocusCallback(focusCallback)
	window.SetCursorPosCallback(mouseCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)
	window.SetScrollCallback(scrollCallback)
	window.SetKeyCallback(keyCallBack)
	glfw.SwapInterval(0)

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	gl.Viewport(0, 0, int32(width), int32(height))
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)
}

// opens and reads shader files, compiles them and returns program
func newProgram(vertexShaderS string, fragmentShaderS string) uint32 {
	v, err := os.Open(vertexShaderS)
	if err != nil {
		log.Fatal(err)
	}
	defer v.Close()

	f, err := os.Open(fragmentShaderS)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	frag := ""
	vert := ""

	scannerV := bufio.NewScanner(v)
	for scannerV.Scan() {
		vert += "\n" + scannerV.Text()
	}
	if err := scannerV.Err(); err != nil {
		log.Fatal(err)
	}
	vert += "\x00"

	scannerF := bufio.NewScanner(f)
	for scannerF.Scan() {
		frag += "\n" + scannerF.Text()
	}
	if err := scannerF.Err(); err != nil {
		log.Fatal(err)
	}
	frag += "\x00"

	vertexShader, err := compileShader(vert, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(frag, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

// compiles a shader
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
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

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

// makeVao initializes and returns a vertex array from points, colors, texture coordinates
func makeVaoEarth(vertices []float32, indices []uint32, stride int32) uint32 {
	var vertexBuffer, elementBuffer, vertexArray uint32

	gl.GenBuffers(1, &vertexBuffer)
	gl.GenBuffers(1, &elementBuffer)
	gl.GenVertexArrays(1, &vertexArray)

	gl.BindVertexArray(vertexArray)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	if indices != nil {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, elementBuffer)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices), gl.Ptr(indices), gl.STATIC_DRAW)
	}

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return vertexArray
}

// generates a vertex array with no texture coordinates
func makeVaoColoredLines(vertices []float32, indices []uint32, stride int32) uint32 {
	var vertexBuffer, elementBuffer, vertexArray uint32

	gl.GenBuffers(1, &vertexBuffer)
	gl.GenBuffers(1, &elementBuffer)
	gl.GenVertexArrays(1, &vertexArray)

	gl.BindVertexArray(vertexArray)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	if indices != nil {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, elementBuffer)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices), gl.Ptr(indices), gl.STATIC_DRAW)
	}

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return vertexArray
}

//generates and returns and OpenGL texture object
func generateTexture(path string) uint32 {
	pixels, x, y := loadImage(path)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.GenerateMipmap(gl.TEXTURE_2D)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, x, y, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	return texture
}

func genGrid(nCols int, nRows int, x1 float64, y1 float64, x2 float64, y2 float64) []float32 {
	vertices := []float32{}

	xStep := math.Abs(x2-x1) / float64(nCols-1)
	yStep := math.Abs(y2-y1) / float64(nRows-1)

	for i := 0; i < nCols; i++ {
		pos1X := x1 + float64(i)*xStep
		pos1Y := y1

		pos2X := x1 + float64(i)*xStep
		pos2Y := y2

		pos1Z := 0.0
		pos2Z := 0.0

		var r, g, b float32 = 1.0, 1.0, 1.0

		vertices = append(vertices, float32(pos1X), float32(pos1Y), float32(pos1Z), r, g, b, float32(pos2X), float32(pos2Y), float32(pos2Z), r, g, b)
	}

	for j := 0; j < nRows; j++ {
		pos1X := x1
		pos1Y := y1 - float64(j)*yStep

		pos2X := x2
		pos2Y := -y2 - float64(j)*yStep

		pos1Z := 0.0
		pos2Z := 0.0

		var r, g, b float32 = 1.0, 1.0, 1.0

		vertices = append(vertices, float32(pos1X), float32(pos1Y), float32(pos1Z), r, g, b, float32(pos2X), float32(pos2Y), float32(pos2Z), r, g, b)
	}

	return vertices
}
