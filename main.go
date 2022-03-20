// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build example
// +build example

package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 240
	screenHeight = 240
	proX         = screenWidth / 2
	proY         = screenHeight / 2

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 8
)
const (
	tileSize = 16
	tileXNum = 25
)

var (
	tilesImage   *ebiten.Image
	textBoxImage *ebiten.Image
)

var (
	runnerImage *ebiten.Image
)

type Position struct {
	X float64
	Y float64
}

type Game struct {
	count               int
	moveCount           int
	protaganistPosition Position
	keys                []ebiten.Key
	layers              [][][]int
	showTextBox         bool
	textBoxLayers       [][][]int
	currText            string
}

func (g *Game) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	g.count += 1
	return nil
}

func modifyProtaganistX(g *Game, screenWidth, step int) {
	newStep := 0
	if len(g.keys) > 0 {
		if g.keys[len(g.keys)-1] == ebiten.KeyLeft {
			newStep = -(step)
		} else if g.keys[len(g.keys)-1] == ebiten.KeyRight {
			newStep = step
		}
	}

	g.protaganistPosition.X = float64(int(g.protaganistPosition.X) + newStep)
}

func modifyProtagainstY(g *Game, screenWidth, step int) {
	newStep := 0
	if len(g.keys) > 0 {
		if g.keys[len(g.keys)-1] == ebiten.KeyUp {
			newStep = -(step)
		} else if g.keys[len(g.keys)-1] == ebiten.KeyDown {
			newStep = step
		}
	}

	g.protaganistPosition.Y = float64(int(g.protaganistPosition.Y) + newStep)
}

func drawProtaganistMain(g *Game, x, y float64, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	i := (g.count / 5) % frameNum
	sx, sy := frameOX+i*frameWidth, frameOY
	screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image), op)
}

func drawTextBox(g *Game, x, y float64, newLayers [][][]int, show bool, screen *ebiten.Image) {
	if !show {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, screenHeight-screenHeight/3)
	screen.DrawImage(textBoxImage, op)
}

func drawBackgroundLayers(g *Game, newLayers [][][]int, screen *ebiten.Image) {
	const xNum = screenWidth / tileSize
	if g.protaganistPosition.X != proX {
		difference := g.protaganistPosition.X - proX
		for layerNum, layer := range newLayers {
			for lineNum, line := range layer {
				for pixelNum := range line {
					// move right
					if difference > 0 {
						newLayers[layerNum][lineNum] = ShiftIntArrRight(line, int(difference))
						// move left
					} else if difference < 0 {
						newLayers[layerNum][lineNum] = ShiftIntArrLeft(line, int(math.Abs(difference)))
					} else {
						newLayers[layerNum][lineNum][pixelNum] = 0
					}
				}
			}
		}
	}
	if g.protaganistPosition.Y != proY {
		difference := g.protaganistPosition.Y - proY
		for layerNum, layer := range newLayers {
			// move up
			if difference > 0 {
				newLayers[layerNum] = ShiftLayerUp(layer, int(difference))
				// move down
			} else if difference < 0 {
				newLayers[layerNum] = ShiftLayerDown(layer, int(math.Abs(difference)))
			}
		}
	}
	for _, sl := range newLayers {
		l := []int{}
		for _, ml := range sl {
			l = append(l, ml...)
		}
		for i, t := range l {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((i%xNum)*tileSize), float64((i/xNum)*tileSize))

			sx := (t % tileXNum) * tileSize
			sy := (t / tileXNum) * tileSize
			image := tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image)
			_ = image
			screen.DrawImage(image, op)
		}
	}
}

func collectUserInput(g *Game, width, height int) {
	modifyProtaganistX(g, width, 1)
	modifyProtagainstY(g, height, 1)
	applyBoundries()
}

func applyBoundries() {

}

func (g *Game) Draw(screen *ebiten.Image) {
	collectUserInput(g, screenWidth, screenHeight)
	drawBackgroundLayers(g, DuplicateLayers(g.layers), screen)
	drawProtaganistMain(g, proX, proY, screen)
	drawTextBox(g, 0, 0, DuplicateLayers(g.textBoxLayers), true, screen)

	// Reset everything
	resetDrawValues(g, screen)
}

func resetDrawValues(g *Game, screen *ebiten.Image) {
	g.keys = []ebiten.Key{}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Decode an image from the image file's byte slice.
	// Now the byte slice is generated with //go:generate for Go 1.15 or older.
	// If you use Go 1.16 or newer, it is strongly recommended to use //go:embed to embed the image file.
	// See https://pkg.go.dev/embed for more details.
	/*
		img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
		if err != nil {
			log.Fatal(err)
		}
		runnerImage = ebiten.NewImageFromImage(img)
	*/
	runnerImage, _, _ = ebitenutil.NewImageFromFile("runner.png")

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("I am not a Monster")
	g := loadGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func loadGame() *Game {
	g := &Game{
		layers: [][][]int{
			{
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 218, 243, 244, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},

				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 244, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 219, 243, 243, 243, 219, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},

				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 244, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
			},
			{
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 26, 27, 28, 29, 30, 31, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 51, 52, 53, 54, 55, 56, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 76, 77, 78, 79, 80, 81, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 101, 102, 103, 104, 105, 106, 0, 0, 0, 0},

				{0, 0, 0, 0, 0, 126, 127, 128, 129, 130, 131, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 303, 303, 245, 242, 303, 303, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},

				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
			},
		},
		textBoxLayers: [][][]int{
			{
				{400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400},
				{400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400},
				{400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400},
				{400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400},
				{400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400},
			},
		},
		protaganistPosition: Position{screenWidth / 2, screenHeight / 2},
	}
	return g
}

func init() {
	tilesImage, _, _ = ebitenutil.NewImageFromFile("tiles.png")
	textBoxImage, _, _ = ebitenutil.NewImageFromFile("textbox_combined.png")
}

/*
layers: [][][]int{
			{
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 218, 243, 244, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},

				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 244, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 219, 243, 243, 243, 219, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},

				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
				{243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 244, 243, 243, 243},
				{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
			},
			{
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 26, 27, 28, 29, 30, 31, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 51, 52, 53, 54, 55, 56, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 76, 77, 78, 79, 80, 81, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 101, 102, 103, 104, 105, 106, 0, 0, 0, 0},

				{0, 0, 0, 0, 0, 126, 127, 128, 129, 130, 131, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 303, 303, 245, 242, 303, 303, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},

				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0},
			},
		},
		protaganistPosition: Position{screenWidth / 2, screenHeight / 2},
	}
*/

// Will move the array a number of slots over right, filling with a specified character
func ShiftIntArrRight(arrToShift []int, spaceToShift int) []int {
	newArr := make([]int, 0)

	if len(arrToShift) < spaceToShift {
		return make([]int, len(arrToShift))
	}

	suppArr := make([]int, spaceToShift)

	newArr = append(newArr, arrToShift[spaceToShift:]...)
	newArr = append(newArr, suppArr...)

	return newArr
}

// Will move the array a number of slots over left, filling with a specified character
func ShiftIntArrLeft(arrToShift []int, spaceToShift int) []int {
	newArr := make([]int, 0)

	if len(arrToShift) < spaceToShift {
		return make([]int, len(arrToShift))
	}

	suppArr := make([]int, spaceToShift)

	newArr = append(newArr, suppArr...)
	newArr = append(newArr, arrToShift[:len(arrToShift)-spaceToShift]...)

	return newArr
}

func ShiftLayerUp(layerToShift [][]int, spaceToShift int) [][]int {
	if len(layerToShift) == 0 {
		return [][]int{}
	}
	newLayer := make([][]int, 0)

	// overflow
	if spaceToShift >= len(layerToShift) {
		newRow := make([]int, len(layerToShift[0]))
		for range layerToShift {
			newLayer = append(newLayer, newRow)
		}
		return newLayer
	}

	newArr := layerToShift[spaceToShift:]
	for i := 0; i < spaceToShift; i++ {
		newRow := make([]int, len(layerToShift[0]))
		newArr = append(newArr, newRow)
	}

	return newArr
}

func ShiftLayerDown(layerToShift [][]int, spaceToShift int) [][]int {
	if len(layerToShift) == 0 {
		return [][]int{}
	}
	newLayer := make([][]int, 0)

	if spaceToShift >= len(layerToShift) {
		newRow := make([]int, len(layerToShift[0]))
		for range layerToShift {
			newLayer = append(newLayer, newRow)
		}
		return newLayer
	}

	newArr := make([][]int, 0)
	for i := 0; i < spaceToShift; i++ {
		newRow := make([]int, len(layerToShift[0]))
		newArr = append(newArr, newRow)
	}
	newArr = append(newArr, layerToShift[:len(layerToShift)-spaceToShift]...)

	return newArr
}

func AreIntArrSame(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for index, val := range a {
		if val != b[index] {
			return false
		}
	}

	return true
}

func AreLayerSame(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}

	for rowIndex, row := range a {
		if !AreIntArrSame(row, b[rowIndex]) {
			return false
		}
	}

	return true
}

func DumpIntArr(a []int) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), ","), "[]")
}

// Dumps a layer (unoptimized)
func DumpLayer(a [][]int) string {
	newString := ""
	for _, row := range a {
		newString += DumpIntArr(row)
		newString += "\n"
	}
	return newString
}

func DuplicateLayers(layers [][][]int) [][][]int {
	newLayers := make([][][]int, len(layers))
	for layerIndex, layer := range layers {
		newLayer := make([][]int, len(layer[layerIndex]))
		for rowIndex, row := range layer {
			newRow := make([]int, len(row))
			copy(newRow, row)
			newLayer[rowIndex] = newRow
		}
		newLayers[layerIndex] = newLayer
	}

	return newLayers
}
