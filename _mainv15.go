package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 800
	gridSize     = 1
	gridWidth    = screenWidth / gridSize
	gridHeight   = screenHeight / gridSize
	waveSpeed    = 0.25
	damping      = 0.995
)

type WaveGrid struct {
	current  [][]float64
	previous [][]float64
	mask     [][]bool // true if inside shape
	shape    []Vector2
}

type Vector2 struct {
	x, y float64
}

func NewWaveGrid() *WaveGrid {
	wg := &WaveGrid{
		current:  make([][]float64, gridHeight),
		previous: make([][]float64, gridHeight),
		mask:     make([][]bool, gridHeight),
		shape:    generateCircleShape(screenWidth/2, screenHeight/2, 150),
	}

	for i := range wg.current {
		wg.current[i] = make([]float64, gridWidth)
		wg.previous[i] = make([]float64, gridWidth)
		wg.mask[i] = make([]bool, gridWidth)
	}

	wg.initializeMask()
	return wg
}

func generateCircleShape(cx, cy, radius float64) []Vector2 {
	var shape []Vector2
	segments := 100
	for i := 0; i < segments; i++ {
		angle := (float64(i) / float64(segments)) * 2 * math.Pi
		x := cx + radius*math.Cos(angle)
		y := cy + radius*math.Sin(angle)
		shape = append(shape, Vector2{x, y})
	}
	return shape
}

func (wg *WaveGrid) initializeMask() {
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			px := float64(x * gridSize)
			py := float64(y * gridSize)
			wg.mask[y][x] = wg.pointInShape(px, py)
		}
	}
}

func (wg *WaveGrid) pointInShape(px, py float64) bool {
	// Center and radius of circle
	cx, cy := screenWidth/2, screenHeight/2
	radius := 150.0

	dx := px - float64(cx)
	dy := py - float64(cy)
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist < radius
}

func (wg *WaveGrid) addWave(mx, my float64) {
	gridX := int(mx / gridSize)
	gridY := int(my / gridSize)

	if gridX >= 0 && gridX < gridWidth && gridY >= 0 && gridY < gridHeight && wg.mask[gridY][gridX] {
		wg.current[gridY][gridX] += 20.0
	}
}

func (wg *WaveGrid) update() {
	// Wave equation: y_new = 2*y_current - y_previous + c^2 * laplacian
	c2 := waveSpeed * waveSpeed

	for y := 1; y < gridHeight-1; y++ {
		for x := 1; x < gridWidth-1; x++ {
			if !wg.mask[y][x] {
				continue
			}

			// Check if neighbors are in shape for proper boundary handling
			laplacian := 0.0
			numNeighbors := 0

			// Check all 4 neighbors
			neighbors := []struct{ dx, dy int }{
				{0, -1}, {0, 1}, {-1, 0}, {1, 0},
			}

			for _, n := range neighbors {
				nx := x + n.dx
				ny := y + n.dy

				if nx >= 0 && nx < gridWidth && ny >= 0 && ny < gridHeight {
					if wg.mask[ny][nx] {
						// Neighbor is in the shape
						laplacian += wg.current[ny][nx]
						numNeighbors++
					} else {
						// Neighbor is a boundary - wave reflects
						laplacian += -wg.current[y][x]
						numNeighbors++
					}
				}
			}

			laplacian -= float64(numNeighbors) * wg.current[y][x]

			// Calculate new height
			newHeight := 2*wg.current[y][x] - wg.previous[y][x] + c2*laplacian
			newHeight *= damping

			wg.previous[y][x] = wg.current[y][x]
			wg.current[y][x] = newHeight
		}
	}

	// Boundary conditions: waves can't escape
	for x := 0; x < gridWidth; x++ {
		wg.current[0][x] = 0
		wg.current[gridHeight-1][x] = 0
	}
	for y := 0; y < gridHeight; y++ {
		wg.current[y][0] = 0
		wg.current[y][gridWidth-1] = 0
	}
}

func (wg *WaveGrid) draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw wave grid
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			if !wg.mask[y][x] {
				continue
			}

			height := wg.current[y][x]

			// Smooth color gradient based on wave height
			normalizedHeight := height / 50.0 // Normalize for color mapping
			normalizedHeight = math.Max(-1, math.Min(1, normalizedHeight))

			var r, g, b uint8

			if normalizedHeight > 0 {
				// Positive wave = gradient from dark to bright blue
				b = uint8(100 + normalizedHeight*155)
				g = uint8(100 + normalizedHeight*80)
				r = uint8(30 + normalizedHeight*30)
			} else {
				// Negative wave = gradient from dark to bright red
				r = uint8(100 - normalizedHeight*155)
				g = uint8(80 - normalizedHeight*80)
				b = uint8(80 - normalizedHeight*50)
			}

			px := float32(x * gridSize)
			py := float32(y * gridSize)
			vector.DrawFilledRect(screen, px, py, float32(gridSize), float32(gridSize), color.RGBA{r, g, b, 255}, false)
		}
	}

	// Draw shape boundary
	if len(wg.shape) > 1 {
		for i := 0; i < len(wg.shape)-1; i++ {
			p1 := wg.shape[i]
			p2 := wg.shape[i+1]
			vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), 3, color.RGBA{255, 200, 100, 255}, false)
		}
		// Close the shape
		p1 := wg.shape[len(wg.shape)-1]
		p2 := wg.shape[0]
		vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), 3, color.RGBA{255, 200, 100, 255}, false)
	}
}

type Game struct {
	waveGrid *WaveGrid
}

func NewGame() *Game {
	return &Game{
		waveGrid: NewWaveGrid(),
	}
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.waveGrid.addWave(float64(x), float64(y))
	}

	g.waveGrid.update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.waveGrid.draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nClick inside the shape to create waves", ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Wave Simulation - Pond")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
