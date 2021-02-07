package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/gdamore/tcell"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/rivo/tview"
)

var selected [][]string
var mutex = &sync.Mutex{}

var root *tview.TreeNode
var app *tview.Application
var m = make(map[string]int)
var kml Folder

// Show a navigable tree view of the current directory.
func gui(win *glfw.Window) {
	kml = readKML(visualOutputPath, 0)
	rootDir := "Document"
	var rootSlice = []string{"true"}
	root = tview.NewTreeNode(rootDir).SetReference(rootSlice).SetSelectable(true)

	tree := tview.NewTreeView().SetGraphicsColor(tcell.NewRGBColor(191, 48, 141)).
		SetRoot(root).
		SetCurrentNode(root)

	optionForm := tview.NewForm().
		AddCheckbox("Show Earth", state.showEarth, showEarthCallback).SetItemPadding(1).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.NewRGBColor(48, 70, 192)).
		SetFieldTextColor(tcell.ColorWhite).
		AddCheckbox("Show Orbits", state.showOrbits, showOrbitsCallback).
		AddCheckbox("Show Points", state.showPoints, showPointsCallback).
		AddCheckbox("Show LOS/Blocked/Basis Lines", state.showLines, showLinesCallback).
		AddCheckbox("Enable Antialiasing (MSAA) (Performance Impact: HIGH)", state.enableAntialiasing, enableAntialiasingCallback).
		AddCheckbox("Enable OpenGL Blending (Performance Impact: MEDIUM)", state.enableBlending, enableBlendingCallback)
	options := tview.NewFlex().
		AddItem(optionForm, 0, 1, true)

	options.SetDirection(tview.FlexRow).
		SetBorder(true).
		SetBackgroundColor(tcell.ColorBlack).
		SetTitle("Options").
		SetBorderColor(tcell.NewRGBColor(191, 48, 141))

	controlBox := tview.NewTextView().SetWordWrap(false).SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	controlBox.SetBorder(true).SetTitle("Controls").SetBorderColor(tcell.NewRGBColor(48, 70, 192)).SetBorderPadding(1, 1, 0, 0)

	word := " [black:#BF308D]             IN TERMINAL              [white] \n" +
		" Select KML.................[#000000:#3046c0]     1     [white] \n" +
		" Select Render Attributes...[#000000:#3046c0]     2     [white] \n" +
		" Move Up Tree...............[#000000:#3046c0]  ArrowUp  [white] \n" +
		" Move Down Tree.............[#000000:#3046c0] ArrowDown [white] \n" +
		" Move Up Tree By Page.......[#000000:#3046c0]  Page_Up  [white] \n" +
		" Move Down Tree by Page.....[#000000:#3046c0] Page_Down [white] \n" +
		" Select/Deselect............[#000000:#3046c0]   Space   [white] \n" +
		" Reload KML.................[#000000:#3046c0]     X     [white] \n" +
		" Collapse Node..............[#000000:#3046c0]     Z     [white] \n" +
		" Show/Hide Controls.........[#000000:#3046c0]     C     [white] \n" +
		" [black:#BF308D]             IN WINDOW                [white] \n" +
		" Move Forward...............[#000000:#3046c0]     W     [white] \n" +
		" Move Backward..............[#000000:#3046c0]     S     [white] \n" +
		" Move Left..................[#000000:#3046c0]     A     [white] \n" +
		" Move Right.................[#000000:#3046c0]     D     [white] \n" +
		" Move Up....................[#000000:#3046c0]   Space   [white] \n" +
		" Move Down..................[#000000:#3046c0]   Shift   [white] \n" +
		" Increase/Decrease Speed....[#000000:#3046c0]   Scroll  [white] \n" +
		" Show/Hide Orbits...........[#000000:#3046c0]     4     [white] \n" +
		" Show/Hide Points...........[#000000:#3046c0]     3     [white] \n" +
		" Show/Hide Lines............[#000000:#3046c0]     2     [white] \n" +
		" Show/Hide Earth............[#000000:#3046c0]     1     [white] \n" +
		" Rotate Earth Left..........[#000000:#3046c0] ArrowLeft [white] \n" +
		" Rotate Earth Right.........[#000000:#3046c0] ArrowRight[white] \n" +
		" Toggle Mouse Lock..........[#000000:#3046c0] MouseLeft [white] \n" +
		" Quit.......................[#000000:#3046c0]     Q     [white]"

	fmt.Fprintf(controlBox, "%s ", word)

	controls := tview.NewFlex().AddItem(controlBox, 0, 1, false)

	flex := tview.NewFlex().AddItem(tree, 0, 1, true).AddItem(options, 0, 1, false) // .AddItem(controls, 43, 1, false) //.SetDrawFunc(flexDrawFunc)

	//add(root, kml.Folders[0].Name)
	for i := range kml.Folders[0].Folders {
		m[kml.Folders[0].Folders[i].Name] = i
		scanFolder(root, kml.Folders[0].Folders[i], "", kml, i)
	}

	//fmt.Println(kml)wdwd

	// fmt.Println(kml.Folders)
	tree.SetSelectedFunc(onSelect)
	tree.SetBorder(true)
	tree.SetBorderColor(tcell.NewRGBColor(48, 70, 192))
	tree.SetTitle("KML Explorer")

	showControls := true

	inputCallBack := func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Key(); key {
		case tcell.KeyUp:
			k := tcell.KeyBacktab
			return tcell.NewEventKey(k, 0, tcell.ModNone)
		case tcell.KeyDown:
			k := tcell.KeyTab
			return tcell.NewEventKey(k, 0, tcell.ModNone)
		case tcell.KeyCtrlC:
			win.SetShouldClose(true)
			k := tcell.KeyCtrlC
			return tcell.NewEventKey(k, 0, tcell.ModNone)
		case tcell.KeyRune:
			switch event.Rune() {
			case '1':
				app.SetFocus(tree)
				//log.Println("!")
			case '2':
				app.SetFocus(optionForm)
			case ' ':
				k := tcell.KeyEnter
				return tcell.NewEventKey(k, 0, tcell.ModNone)
			case 'q':
				win.SetShouldClose(true)
				k := tcell.KeyCtrlC
				return tcell.NewEventKey(k, 0, tcell.ModNone)
			case 'x':
				reloadKML()
			case 'z':
				n := tree.GetCurrentNode()
				n.SetExpanded(!n.IsExpanded())
			case 'c':
				if showControls {
					flex.RemoveItem(controls)
					showControls = !showControls
				} else {
					flex.AddItem(controls, 43, 1, false)
					showControls = !showControls
				}

			}

		}
		return event
	}

	app = tview.NewApplication().SetRoot(flex, true)
	app.SetInputCapture(inputCallBack)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func add(target *tview.TreeNode, name string, reference string, folder Folder) *tview.TreeNode {
	s := strings.Split(reference, "$")
	s2 := s[2:]
	s2 = append(s2, "t")

	node := tview.NewTreeNode(name[:len(name)-32]).SetReference(s2)

	target.AddChild(node)
	children := target.GetChildren()

	return children[len(children)-1]
}

func scanFolder(node *tview.TreeNode, folder Folder, prefix string, previous Folder, index int) {
	rand := randomString(32)
	folder.Name = folder.Name + rand
	prefix += "$" + previous.Name
	currentNode := add(node, folder.Name, prefix+"$"+folder.Name, folder)
	m[folder.Name] = index
	for i := range folder.Folders {
		//index++
		scanFolder(currentNode, folder.Folders[i], prefix, folder, i)
		//index--
	}
}

func scanTree(node *tview.TreeNode) {
	node.Walk(walkCallback)
}

func walkCallback(current *tview.TreeNode, parent *tview.TreeNode) bool {
	ref := current.GetReference().([]string)
	if ref[len(ref)-1] == "f" && len(current.GetChildren()) == 0 {
		mutex.Lock()
		selected = append(selected, ref[:len(ref)-1])
		mutex.Unlock()
	}
	return true
}

func onSelect(node *tview.TreeNode) {
	//node.SetExpanded(!node.IsExpanded())
	ref := node.GetReference().([]string)

	if ref[len(ref)-1] == "t" {
		setColor(node, tcell.NewRGBColor(48, 70, 192), "f")
	} else if ref[len(ref)-1] == "f" {
		setColor(node, tcell.ColorWhite, "t")
	}

	reloadKML()
}

func setColor(node *tview.TreeNode, color tcell.Color, ref string) {
	node.SetColor(color)
	reference := node.GetReference().([]string)
	reference[len(reference)-1] = ref
	node.SetReference(reference)
	children := node.GetChildren()
	for i := range children {
		setColor(children[i], color, ref)
	}
}

func showEarthCallback(x bool) {
	state.showEarth = x
}

func showOrbitsCallback(x bool) {
	state.showOrbits = x
}

func showLinesCallback(x bool) {
	state.showLines = x
}

func showPointsCallback(x bool) {
	state.showPoints = x
}

func reloadKML() {
	mutex.Lock()
	selected = [][]string{{}}
	mutex.Unlock()
	scanTree(root)
	state.resetting = true
}

func enableAntialiasingCallback(x bool) {
	state.enableAntialiasing = x
}

func enableBlendingCallback(x bool) {
	state.enableBlending = x
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
