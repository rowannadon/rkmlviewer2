package main

import (
	"math"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func processInput(window *glfw.Window, deltaTime float64) {
	if window.GetKey(glfw.KeyQ) == glfw.Press {
		window.SetShouldClose(true)
		app.Stop()
	}

	cameraSpeed := float64(moveSpeed) * deltaTime
	if window.GetKey(glfw.KeyW) == glfw.Press {
		camera.Pos = camera.Pos.Add(camera.Front.Mul(float32(cameraSpeed)))
	}

	if window.GetKey(glfw.KeyS) == glfw.Press {
		camera.Pos = camera.Pos.Sub(camera.Front.Mul(float32(cameraSpeed)))
	}

	if window.GetKey(glfw.KeySpace) == glfw.Press {
		camera.Pos = camera.Pos.Add(camera.Up.Mul(float32(cameraSpeed)))
	}

	if window.GetKey(glfw.KeyLeftShift) == glfw.Press {
		camera.Pos = camera.Pos.Sub(camera.Up.Mul(float32(cameraSpeed)))
	}

	if window.GetKey(glfw.KeyA) == glfw.Press {
		camera.Pos = camera.Pos.Sub(camera.Front.Cross(camera.Up).Normalize().Mul(float32(cameraSpeed)))
	}

	if window.GetKey(glfw.KeyD) == glfw.Press {
		camera.Pos = camera.Pos.Add(camera.Front.Cross(camera.Up).Normalize().Mul(float32(cameraSpeed)))
	}

	if window.GetKey(glfw.KeyLeft) == glfw.Press {
		angleZ -= 0.4
	}

	if window.GetKey(glfw.KeyRight) == glfw.Press {
		angleZ += 0.4
	}
}

func keyCallBack(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.Key1 && action == glfw.Press {
		state.showEarth = !state.showEarth
	}
	if key == glfw.Key2 && action == glfw.Press {
		state.showLines = !state.showLines
	}
	if key == glfw.Key3 && action == glfw.Press {
		state.showPoints = !state.showPoints
	}
	if key == glfw.Key4 && action == glfw.Press {
		state.showOrbits = !state.showOrbits
	}
	if key == glfw.KeyEscape && action == glfw.Press {
		state.inputting = !state.inputting
	}
	if state.inputting {
		mouse.firstMouse = true
		window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}

func mouseCallback(window *glfw.Window, xpos float64, ypos float64) {
	if state.inputting {
		if mouse.firstMouse {
			mouse.lastX = float32(xpos)
			mouse.lastY = float32(ypos)
			mouse.firstMouse = false
		}

		xoffset := float32(xpos) - mouse.lastX
		yoffset := mouse.lastY - float32(ypos)
		mouse.lastX = float32(xpos)
		mouse.lastY = float32(ypos)

		var sensitivity float32
		sensitivity = 0.03
		xoffset *= sensitivity
		yoffset *= sensitivity

		mouse.yaw += xoffset
		mouse.pitch += yoffset

		if mouse.pitch > 89.0 {
			mouse.pitch = 89.0
		}
		if mouse.pitch < -89.0 {
			mouse.pitch = -89.0
		}

		var front mgl32.Vec3
		front[0] = float32(math.Cos(float64(mgl32.DegToRad(mouse.yaw))) * math.Cos(float64(mgl32.DegToRad(mouse.pitch))))
		front[1] = float32(math.Sin(float64(mgl32.DegToRad(mouse.pitch))))
		front[2] = float32(math.Sin(float64(mgl32.DegToRad(mouse.yaw))) * math.Cos(float64(mgl32.DegToRad(mouse.pitch))))
		camera.Front = front.Normalize()
	}
}

func scrollCallback(window *glfw.Window, xoffset float64, yoffset float64) {
	if moveSpeed >= 0.01 && moveSpeed <= 10.0 {
		moveSpeed += float32(yoffset / 4)
	}

	if moveSpeed <= 0.01 {
		moveSpeed = 0.01
	}

	if moveSpeed >= 10.0 {
		moveSpeed = 10.0
	}

	//fmt.Println(moveSpeed)
}

func fbCallback(window *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func mouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press {
		state.inputting = !state.inputting
	}
	if state.inputting {
		mouse.firstMouse = true
		window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}

func focusCallback(window *glfw.Window, z bool) {
	if !z {
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		state.inputting = false
	}
}
