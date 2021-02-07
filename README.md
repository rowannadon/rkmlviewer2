# rkmlviewer

rkmlviewer is an application designed to read and display data encoded in the Keyhole Markup Language.

The application is built to display "large scale" data such as satellite orbits (not suitable for resolution > 100km)

rkmlviewer is written in Go, uses OpenGL 3.3 for rendering, and GLFW for cross platform windowing support.

rkmlviewer is developed on Ubuntu, and has not been tested on other operating systems.

rkmlviewer contains a command line (ncurses style) gui to select application options at runtime.

# Get started
## Clone the repository

* ```hg clone ssh://diorama.lanl.gov//opt/hg/rkmlviewer2```

### "go get" dependencies

* ```go get github.com/go-gl/gl/v3.2-core/gl```
* ```go get github.com/gdamore/tcell```
* ```go get github.com/rivo/tview```
* ```go get github.com/go-gl/glfw/v3.2/glfw```
* ```go get github.com/go-gl/mathgl/mgl32```


### Build an executable from the source code

* The computer's graphics hardware must be recent enough to support OpenGL 3.3 (released in the last 5 years or so)
* ```cd rkmlviewer2/cmd/rkmlviewer/```
* ```go build main.go sphere.go texture.go gui.go kml.go input.go```

### OR run the pre-built executable

* ```cd rkmlviewer2/cmd/rkmlviewer/```
* ```./main <optional command line arguments>```

# Configuration

## Performance options (can be changed at runtime)

* Enable Antialiasing (MSAA): Makes edges appear smoother by sampling each pixel multiple times and then interpolating. This option decreases performance significantly beacuse each pixel must be sampled multiple times by the fragment shader.
* Enable OpenGL Blending: Enables the use of transparent textures. This option must be enabled for clouds to appear and for lines to appear transparent. This option decreases performance.

## Optional command line flags (cannot be changed at runtime)

* ```-file``` - Specifies the .kml file to be read from (default: "../../examples/diorama-visual-output.kml")
* ```-fov``` - Sets the field of view of the camera in degrees (default: 50, range: 1-179). Higher resolutions may need an increased field of view to appear natural.
* ```-width``` - Sets the width of the window (default: 800)
* ```-height``` - Sets the height of the window (default: 600)
* ```-ps``` - Sets the size (in pixels) that points are drawn at. (default: 8.0)
* ```-grid``` - Draws a grid (default: false, range: true,false)
* ```-ambient``` - Sets the level of global illumination in the scene (default: 0.2, range: 0.0-1.0). A higher value increases ambient light.
* ```-alpha``` - Sets the transparency of the lines when OpenGL blending is enabled (default: 0.6). Higher values are more opaque.
* ```-samples``` Sets the number of samples used by MSAA (default 8, range: 2-16). More samples produces smoother lines at the cost of performance.
* ```-modelres``` - Sets the resolution multiplier for the earth model (default: 4, range: 1-16). Higher resolutions make the edges of the earth appear smoother at the cost of performance.

## Control list

### In the viewing window

* Lock/Unlock Mouse to Window: ```Esc``` or ```MouseLeft```
* Move Forward: ```W```
* Move Backward: ```S```
* Move Left: ```A```
* Move Right: ```D```
* Move Up: ```Space```
* Move Down: ```Shift```
* Increase/Decrease Movement Speed: ```Scroll wheel```
* Show/Hide Earth: ```1```
* Show/Hide Lines: ```2```
* Show/Hide Points: ```3```
* Show/Hide Satellite Orbits: ```4```
* Rotate Earth Left: ```ArrowLeft```
* Rotate Earth Right: ```ArrowRight```
* Quit Application: ```Q```

### In the terminal gui

* Show/Hide Control List: ```C```
* Move Selection Up: ```ArrowUp```
* Move Selection Down: ```ArrowDown```
* Move Selection Up By Page: ```PageUp```
* Move Selection Down By Page: ```PageDown```
* Select/Deselect Option: ```Space``` or ```Enter```
* Collapse/Expand Tree Node: ```Z```
* Reload Selection (should be done automatically): ```X```
* Select 1st Window (KML Explorer): ```1```
* Select 2nd Window (Render Attributes): ```2```
* Quit: ```Q```

## Source organization

###  Contains application source code: ```rkmlviewer/cmd/rkmlviewer```

* ```main.go```
* * Contains main application code
* * Main function: Initiation of OpenGL environment, Main render loop
* * Function to read GLSL shader files, compile them, set uniform in shader program
* * Functions to generate OpenGL vertex array objects
* * Function to generate OpenGL texture object
* ```gui.go```
* * Contains code related to the terminal "GUI"
* * GUI function builds and draws GUI, handles terminal input
* * Functions to recursively build GUI tree based on kml object (generated in ```kml.go```)
* * Functions to handle selection of tree nodes by the user
* * Random string function
* ```input.go```
* * GLFW input callbacks
* ```kml.go```
* * Reads KML into kml object
* * Function to generate array of vertices from selection data (generated in ```gui.go```)
* * Function to generate vertex data from kml objects
* * Function to return color information from KML style
* ```sphere.go```
* * Function to generate sphere vertices
* * Function to convert (lat, lon) to (x, y, z) (origin at center of earth)
* ```texture.go```
* * Function to read image data (jpg, png)

### Contains assets (textures, shaders) required for the application to run: ```rkmlviewer/assets/```

* ```textures/```
* * ```earth.jpg``` Diffuse earth texture
* * ```earth2.jpg``` Alternate diffuse earth texture
* * ```clouds.jpg``` Cloud texture
* * ```night_lights.jpg``` Earth lights at night
* * ```spec_map``` Earth/Ocean mask used to control specular intensity
* ```shaders/```
* * ```vertexshader.glslv``` Vertex shader for earth
* * ```fragmentshader.glslf``` Fragment shader for earth (lighting, atmosphere)
* * ```objectvertexshader``` Vertex shader for line drawing
* * ```objectfragmentshader``` Fragment shader for line drawing (color)
* * ```cloudvertexshader``` Vertex shader for cloud rendering
* * ```cloudfragmentshader``` Fragment shader for cloud rendering (transparency)

### Contains 3rd party dependencies: ```rkmlviewer/vendor/```

### Contains example kml scenarios generated by DIORAMA: ```rkmlviewer/examples/```

* ```diorama.kml``` Very simple scenario, 4 satellites, 1 event
* ```diorama-visual-output.kml``` Complex scenario, many satellites, 1 event
* ```diorama-visual-output.kml``` Very complex scenario, many satellites, 272 events
