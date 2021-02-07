package main

import (
	"math"
)

func generateSphere(sectorCount int, stackCount int, radius float64) ([]float32, []uint32) {
	var PI float32
	PI = 3.1415926
	var vertices []float32
	//fmt.Println(sectorCount, stackCount, radius)

	var x, y, z, xy float32
	var nx, ny, nz float32
	var lengthInv = float32(1.0 / radius)
	var s, t float32

	var sectorStep = 2 * PI / float32(sectorCount)
	var stackStep = PI / float32(stackCount)
	var sectorAngle, stackAngle float32

	for i := 0; i <= stackCount; i++ {
		stackAngle = PI/2 - float32(i)*stackStep             // starting from pi/2 to -pi/2
		xy = float32(radius * math.Cos(float64(stackAngle))) // r * cos(u)
		z = float32(radius * math.Sin(float64(stackAngle)))  // r * sin(u)

		// add (sectorCount+1) vertices per stack
		// the first and last vertices have same position and normal, but different tex coords
		for j := 0; j <= sectorCount; j++ {
			sectorAngle = float32(j) * sectorStep // starting from 0 to 2pi

			// vertex position (x, y, z)
			x = xy * float32(math.Cos(float64(sectorAngle))) // r * cos(u) * cos(v)
			y = xy * float32(math.Sin(float64(sectorAngle))) // r * cos(u) * sin(v)
			vertices = append(vertices, x)
			vertices = append(vertices, y)
			vertices = append(vertices, z)

			// normalized vertex normal (nx, ny, nz)
			nx = x * lengthInv
			ny = y * lengthInv
			nz = z * lengthInv
			vertices = append(vertices, nx)
			vertices = append(vertices, ny)
			vertices = append(vertices, nz)

			// vertex tex coord (s, t) range between [0, 1]
			s = float32(j) / float32(sectorCount)
			t = float32(i) / float32(stackCount)
			vertices = append(vertices, s)
			vertices = append(vertices, t)
		}
	}

	var k1, k2 int
	var indices []uint32

	for i := 0; i < stackCount; i++ {
		k1 = i * (sectorCount + 1) // beginning of current stack
		k2 = k1 + sectorCount + 1  // beginning of next stack

		for j := 0; j < sectorCount; j++ {
			// 2 triangles per sector excluding 1st and last stacks
			if i != 0 {
				indices = append(indices, uint32(k1))
				indices = append(indices, uint32(k2))
				indices = append(indices, uint32(k1+1))
			}

			if i != (stackCount - 1) {
				indices = append(indices, uint32(k1+1))
				indices = append(indices, uint32(k2))
				indices = append(indices, uint32(k2+1))
			}
			k1++
			k2++
		}
	}
	return vertices, indices
}

func latLonToVertex(lat float64, lon float64, h float64) (float32, float32, float32) {
	latRad := lat * (math.Pi / 180)
	lonRad := lon * (math.Pi / 180)

	R := h + a

	X := R * math.Cos(latRad) * math.Cos(lonRad)

	Y := R * math.Cos(latRad) * math.Sin(lonRad)

	Z := R * math.Sin(latRad)

	return float32(X) / a, float32(Y) / a, float32(Z) / a
}
