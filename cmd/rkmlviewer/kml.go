package main

import (
	"encoding/xml"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// Folder is basic xml group structure
type Folder struct {
	XMLName     xml.Name   `xml:"Folder"`
	Name        string     `xml:"name"`
	Folders     []Folder   `xml:"Folder"`
	Visibility  bool       `xml:"visibility"`
	StyleURL    string     `xml:"styleUrl"`
	Point       Point      `xml:"Point"`
	Description string     `xml:"description"`
	LineString  LineString `xml:"LineString"`
	Track       Track      `xml:"Track"`
	ID          int
}

// Point has coordinate, altitude data
type Point struct {
	AltitudeMode string `xml:"altitudeMode"`
	Coordinates  string `xml:"coordinates"`
}

// LineString has two coords, altitudemode, etc
type LineString struct {
	AltitudeMode string `xml:"altitudeMode"`
	Coordinates  string `xml:"coordinates"`
	Tessellate   string `xml:"tessellate"`
	Extrude      string `xml:"extrude"`
}

// Track contains coords and times
type Track struct {
	AltitudeMode string   `xml:"altitudeMode"`
	Whens        []string `xml:"when"`
	Coords       []string `xml:"coord"`
}

func readKML(filename string, eventIndex int) Folder {
	// load the KML document

	// read our opened xmlFile as a byte array.
	old, _ := ioutil.ReadFile(filename)

	newXML := strings.Replace(string(old), "Placemark", "Folder", -1)
	newXML = strings.Replace(string(newXML), "Document", "Folder", -1)
	newXML = strings.Replace(string(newXML), "kml", "Folder", -1)

	byteValue := []byte(newXML)

	// we initialize our Users array
	var kml Folder
	//fmt.Println(byteValue)
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &kml)

	return kml
}

func interpretSelected() ([]float32, int, int) {
	vertices := []float32{}
	points := []float32{}
	orbices := []float32{}
	//app.Stop()
	if len(selected) == 0 {
		return vertices, 0, 0
	}

	doc := kml.Folders[0]

	mutex.Lock()

	for i := range selected {
		switch len(selected[i]) {
		case 1:
			//fmt.Println(doc.Folders[m[selected[i][0]]].Name)
			verts, ponts, orbs := appendVert(doc.Folders[m[selected[i][0]]])
			vertices = append(vertices, verts...)
			points = append(points, ponts...)
			orbices = append(orbices, orbs...)
		case 2:
			//fmt.Println(doc.Folders[m[selected[i][0]]].Folders[m[selected[i][1]]].Name)
			verts, ponts, orbs := appendVert(doc.Folders[m[selected[i][0]]].Folders[m[selected[i][1]]])
			vertices = append(vertices, verts...)
			points = append(points, ponts...)
			orbices = append(orbices, orbs...)
		case 3:
			//fmt.Println(doc.Folders[m[selected[i][0]]].Folders[m[selected[i][1]]].Folders[m[selected[i][2]]].Name)
			verts, ponts, orbs := appendVert(doc.Folders[m[selected[i][0]]].Folders[m[selected[i][1]]].Folders[m[selected[i][2]]])
			vertices = append(vertices, verts...)
			points = append(points, ponts...)
			orbices = append(orbices, orbs...)
		case 4:
			//fmt.Println(doc.Folders[m[selected[i][0]]].Folders[m[selected[i][1]]].Folders[m[selected[i][2]]].Folders[m[selected[i][3]]].Name)
			verts, ponts, orbs := appendVert(doc.Folders[m[selected[i][0]]].Folders[m[selected[i][1]]].Folders[m[selected[i][2]]].Folders[m[selected[i][3]]])
			vertices = append(vertices, verts...)
			points = append(points, ponts...)
			orbices = append(orbices, orbs...)
		}
	}

	mutex.Unlock()

	pointStart := len(vertices)

	vertices = append(vertices, points...)

	orbitStart := len(vertices)

	vertices = append(vertices, orbices...)

	//app.Stop()

	//fmt.Println(m)

	return vertices, pointStart, orbitStart
}

func appendVert(f Folder) ([]float32, []float32, []float32) {
	s := f.LineString.Coordinates
	st := f.StyleURL
	so := f.Track.Coords
	vertices := []float32{}
	points := []float32{}
	orbitVertices := []float32{}
	sp := f.Point.Coordinates

	if len(s) > 0 {
		pos1 := strings.Split(strings.Fields(s)[0], ",")
		pos2 := strings.Split(strings.Fields(s)[1], ",")

		pos1Lat, _ := strconv.ParseFloat(pos1[0], 64)
		pos1Lon, _ := strconv.ParseFloat(pos1[1], 64)
		pos1Alt, _ := strconv.ParseFloat(pos1[2], 64)

		pos2Lat, _ := strconv.ParseFloat(pos2[0], 64)
		pos2Lon, _ := strconv.ParseFloat(pos2[1], 64)
		pos2Alt, _ := strconv.ParseFloat(pos2[2], 64)

		pos1X, pos1Y, pos1Z := latLonToVertex(pos1Lon, pos1Lat, pos1Alt)
		pos2X, pos2Y, pos2Z := latLonToVertex(pos2Lon, pos2Lat, pos2Alt)

		r, g, b := getColor(st)

		vertices = append(vertices, pos1X, pos1Y, pos1Z, r, g, b, pos2X, pos2Y, pos2Z, r, g, b)
	}

	if len(sp) > 0 {
		pos1 := strings.Split(strings.Fields(sp)[0], ",")

		pos1Lat, _ := strconv.ParseFloat(pos1[0], 64)
		pos1Lon, _ := strconv.ParseFloat(pos1[1], 64)
		pos1Alt, _ := strconv.ParseFloat(pos1[2], 64)

		pos1X, pos1Y, pos1Z := latLonToVertex(pos1Lon, pos1Lat-45.492, pos1Alt)

		r, g, b := getColor(st)

		points = append(points, pos1X, pos1Y, pos1Z, r, g, b)
	}

	if len(so) > 0 {
		r, g, b := getColor(st)
		for i := range so {
			pos1 := strings.Split(strings.Fields(so[i])[0], ",")

			pos1Lat, _ := strconv.ParseFloat(pos1[0], 64)
			pos1Lon, _ := strconv.ParseFloat(pos1[1], 64)
			pos1Alt, _ := strconv.ParseFloat(pos1[2], 64)

			pos1X, pos1Y, pos1Z := latLonToVertex(pos1Lon, pos1Lat, pos1Alt)

			if i > 0 {
				orbitVertices = append(orbitVertices, pos1X, pos1Y, pos1Z, r, g, b)
			}
			orbitVertices = append(orbitVertices, pos1X, pos1Y, pos1Z, r, g, b)
		}

		pos1 := strings.Split(strings.Fields(so[0])[0], ",")

		pos1Lat, _ := strconv.ParseFloat(pos1[0], 64)
		pos1Lon, _ := strconv.ParseFloat(pos1[1], 64)
		pos1Alt, _ := strconv.ParseFloat(pos1[2], 64)

		pos1X, pos1Y, pos1Z := latLonToVertex(pos1Lon, pos1Lat, pos1Alt)

		orbitVertices = append(orbitVertices, pos1X, pos1Y, pos1Z, r, g, b)
	}

	return vertices, points, orbitVertices
}

func getColor(color string) (float32, float32, float32) {
	switch color {
	case "red_line":
		return 1.0, 0.0, 0.0
	case "green_line":
		return 0.0, 1.0, 0.0
	case "blue_line":
		return 0.0, 0.0, 1.0
	case "pink_line":
		return 1.0, 0.0, 1.0
	case "cyan_line":
		return 0.0, 1.0, 1.0
	case "purple_line":
		return 0.5, 0.0, 1.0
	case "orange_line":
		return 1.0, 0.5, 0.0
	case "yellow_line":
		return 1.0, 1.0, 0.0
	case "#webcam1":
		return rand.Float32(), 1.0, float32(math.Max(rand.Float64(), 0.5))
	case "shaded_dot":
		return 1.0, 0.0, 0.5
	}
	return 1.0, 1.0, 1.0
}
