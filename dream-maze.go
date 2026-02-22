package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

///////////////////////////////////////////////////////////////
// GLOBALS
///////////////////////////////////////////////////////////////

var currentSize = 31
var gameMap [][]int // 0 empty, 1 wall, 2 exit

var levelMessages = map[int]string{
	1: "Brain condition is still critical...",
	2: "Brain condition is getting worse...",
	3: "Risk of brain death is gone.",
	4: "...",
}

///////////////////////////////////////////////////////////////
// PLAYER
///////////////////////////////////////////////////////////////

type Player struct {
	x, y  float64
	angle float64
	speed float64
}

///////////////////////////////////////////////////////////////
// GAME
///////////////////////////////////////////////////////////////

type Game struct {
	player   Player
	level    int
	finished bool
}

///////////////////////////////////////////////////////////////
// MAZE GENERATION
///////////////////////////////////////////////////////////////

func generateMaze(level int) {

	gameMap = make([][]int, currentSize)
	for i := range gameMap {
		gameMap[i] = make([]int, currentSize)
	}

	for y := 0; y < currentSize; y++ {
		for x := 0; x < currentSize; x++ {
			gameMap[y][x] = 1
		}
	}

	type cell struct{ x, y int }
	stack := []cell{{1, 1}}
	gameMap[1][1] = 0
	rand.Seed(time.Now().UnixNano())

	dirs := []cell{{0, -2}, {2, 0}, {0, 2}, {-2, 0}}

	// Perfect DFS maze
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })

		for _, d := range dirs {
			nx := current.x + d.x
			ny := current.y + d.y
			if nx > 0 && nx < currentSize-1 && ny > 0 && ny < currentSize-1 && gameMap[ny][nx] == 1 {
				gameMap[ny][nx] = 0
				gameMap[current.y+d.y/2][current.x+d.x/2] = 0
				stack = append(stack, cell{nx, ny})
			}
		}
	}

	// Chaos Phase
	var chaosRatio float64
	var minOpenNeighbors int

	switch level {
	case 1:
		chaosRatio = 0.2
		minOpenNeighbors = 2
	case 2:
		chaosRatio = 0.6
		minOpenNeighbors = 1
	case 3:
		chaosRatio = 0.1
		minOpenNeighbors = 2
	case 4:
		chaosRatio = 0.0
	}

	chaosCount := int(float64(currentSize*currentSize) * chaosRatio)
	for i := 0; i < chaosCount; i++ {
		x := rand.Intn(currentSize-2) + 1
		y := rand.Intn(currentSize-2) + 1
		if gameMap[y][x] == 1 {
			openCount := 0
			for _, d := range dirs {
				nx, ny := x+d.x, y+d.y
				if nx > 0 && nx < currentSize-1 && ny > 0 && ny < currentSize-1 {
					if gameMap[ny][nx] == 0 {
						openCount++
					}
				}
			}
			if openCount >= minOpenNeighbors {
				gameMap[y][x] = 0
			}
		}
	}
}

///////////////////////////////////////////////////////////////
// EXIT
///////////////////////////////////////////////////////////////

func placeExit(px, py int) {
	var maxDist float64
	var ex, ey int

	for y := 1; y < currentSize-1; y++ {
		for x := 1; x < currentSize-1; x++ {
			if gameMap[y][x] == 0 {
				d := math.Hypot(float64(px-x), float64(py-y))
				if d > maxDist {
					maxDist = d
					ex = x
					ey = y
				}
			}
		}
	}

	gameMap[ey][ex] = 2
}

///////////////////////////////////////////////////////////////
// UPDATE
///////////////////////////////////////////////////////////////

func (g *Game) Update() error {

	if g.finished {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.level = 1
			g.finished = false
			currentSize = 31
			g.generateLevel()
		}
		return nil
	}

	p := &g.player

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		nx := p.x + math.Cos(p.angle)*p.speed
		ny := p.y + math.Sin(p.angle)*p.speed
		if gameMap[int(ny)][int(nx)] != 1 {
			p.x = nx
			p.y = ny
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		nx := p.x - math.Cos(p.angle)*p.speed
		ny := p.y - math.Sin(p.angle)*p.speed
		if gameMap[int(ny)][int(nx)] != 1 {
			p.x = nx
			p.y = ny
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.angle -= 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.angle += 0.05
	}

	// Exit reached
	if gameMap[int(p.y)][int(p.x)] == 2 {
		if g.level == 4 {
			g.finished = true
			return nil
		}
		g.level++
		currentSize += 12
		if currentSize%2 == 0 {
			currentSize++
		}
		g.generateLevel()
		p.x = 1.5
		p.y = 1.5
		p.angle = 0
	}

	return nil
}

func (g *Game) generateLevel() {
	generateMaze(g.level)
	placeExit(1, 1)
}

///////////////////////////////////////////////////////////////
// DRAW
///////////////////////////////////////////////////////////////

func (g *Game) Draw(screen *ebiten.Image) {

	width, height := screen.Size()

	if g.finished {
		screen.Fill(color.RGBA{10, 10, 15, 255})
		// Ground
		ebitenutil.DrawRect(screen, 0, 170, 320, 30, color.RGBA{40, 40, 40, 255})
		// Rails
		ebitenutil.DrawLine(screen, 0, 180, 320, 180, color.RGBA{120, 120, 120, 255})
		ebitenutil.DrawLine(screen, 0, 190, 320, 190, color.RGBA{120, 120, 120, 255})
		// Smoke
		ebitenutil.DrawRect(screen, 150, 80, 10, 10, color.RGBA{200, 200, 200, 200})
		ebitenutil.DrawRect(screen, 155, 70, 12, 12, color.RGBA{180, 180, 180, 180})
		ebitenutil.DrawRect(screen, 160, 55, 14, 14, color.RGBA{160, 160, 160, 160})
		// Engine
		ebitenutil.DrawRect(screen, 120, 120, 70, 35, color.RGBA{170, 0, 0, 255})
		ebitenutil.DrawRect(screen, 100, 130, 25, 25, color.RGBA{140, 0, 0, 255})
		ebitenutil.DrawRect(screen, 155, 100, 35, 25, color.RGBA{120, 0, 0, 255})
		ebitenutil.DrawRect(screen, 165, 105, 15, 12, color.RGBA{100, 200, 255, 255})
		ebitenutil.DrawRect(screen, 135, 95, 12, 25, color.RGBA{60, 60, 60, 255})

		for i := 0; i < 3; i++ {
			ebitenutil.DrawRect(screen, 115+30*float64(i), 155, 22, 22, color.Black)
			ebitenutil.DrawRect(screen, 120+30*float64(i), 160, 12, 12, color.RGBA{100, 100, 100, 255})
		}
		// Wagon
		ebitenutil.DrawRect(screen, 205, 120, 75, 35, color.RGBA{0, 80, 180, 255})
		for i := 0; i < 2; i++ {
			ebitenutil.DrawRect(screen, 215+35*float64(i), 155, 22, 22, color.Black)
		}
		ebitenutil.DebugPrintAt(screen, "You are still alive, Terry.", 95, 20)
		ebitenutil.DebugPrintAt(screen, "Press R to restart", 105, 35)
		return
	}

	// Raycasting
	fov := math.Pi / 3
	for i := 0; i < width; i++ {
		rayAngle := g.player.angle - fov/2 + fov*float64(i)/float64(width)
		distance, cell := castRay(g.player.x, g.player.y, rayAngle)
		corrected := distance * math.Cos(rayAngle-g.player.angle)
		lineHeight := int(float64(height) / corrected)
		col := getWallColor(cell, distance, g.level)
		ebitenutil.DrawLine(screen, float64(i), float64(height/2-lineHeight/2), float64(i), float64(height/2+lineHeight/2), col)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d", g.level), 10, 10)
	if msg, ok := levelMessages[g.level]; ok {
		ebitenutil.DebugPrintAt(screen, msg, 10, 25)
	}
}

///////////////////////////////////////////////////////////////

func castRay(px, py, angle float64) (float64, int) {
	step := 0.05
	distance := 0.0
	var cell int
	for {
		x := px + math.Cos(angle)*distance
		y := py + math.Sin(angle)*distance
		if int(x) < 0 || int(x) >= currentSize || int(y) < 0 || int(y) >= currentSize {
			break
		}
		cell = gameMap[int(y)][int(x)]
		if cell == 1 || cell == 2 {
			break
		}
		distance += step
	}
	if distance == 0 {
		distance = 0.1
	}
	return distance, cell
}

func getWallColor(cell int, distance float64, level int) color.RGBA {
	shade := uint8(200 / (1 + distance*0.1))
	switch level {
	case 1:
		return color.RGBA{shade - 10, shade, shade + 20, 255}
	case 2:
		flicker := uint8(rand.Intn(20))
		return color.RGBA{shade + flicker, shade - 40, shade - 30, 255}
	case 3:
		return color.RGBA{shade + 10, shade + 10, shade, 255}
	case 4:
		return color.RGBA{uint8(float64(shade) * 0.8), uint8(float64(shade) * 1.1), uint8(float64(shade) * 0.6), 255}
	}
	return color.RGBA{shade, shade, shade, 255}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 200
}

///////////////////////////////////////////////////////////////

func main() {
	generateMaze(1)
	placeExit(1, 1)

	game := &Game{player: Player{1.5, 1.5, 0, 0.1}, level: 1}
	ebiten.SetWindowSize(640, 400)
	ebiten.SetWindowTitle("Dream-Maze")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
