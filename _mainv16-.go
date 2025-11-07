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
	waveSpeed    = 1
	damping      = 1 // 0.98
)

type WaveGrid struct {
	height   [][]float64
	velocity [][]float64
	mask     [][]bool
	shape    []Vector2
	cx, cy   float64
	radius   float64
}

type Vector2 struct {
	x, y float64
}

func NewWaveGrid() *WaveGrid {
	wg := &WaveGrid{
		height:   make([][]float64, gridHeight),
		velocity: make([][]float64, gridHeight),
		mask:     make([][]bool, gridHeight),
		cx:       float64(screenWidth) / 2,
		cy:       float64(screenHeight) / 2,
		radius:   150.0,
		shape:    generateCircleShape(screenWidth/2, screenHeight/2, 150),
	}

	for i := range wg.height {
		wg.height[i] = make([]float64, gridWidth)
		wg.velocity[i] = make([]float64, gridWidth)
		wg.mask[i] = make([]bool, gridWidth)
	}

	wg.initializeMask()
	return wg
}

func generateCircleShape(cx, cy, radius float64) []Vector2 {
	var shape []Vector2
	segments := 200
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
			dx := float64(x) - wg.cx
			dy := float64(y) - wg.cy
			dist := math.Sqrt(dx*dx + dy*dy)
			wg.mask[y][x] = dist < wg.radius
		}
	}
}

func (wg *WaveGrid) addWave(mx, my float64) {
	gridX := int(mx)
	gridY := int(my)

	// Add impulse with smooth falloff
	radius := 8.0
	for dy := -int(radius); dy <= int(radius); dy++ {
		for dx := -int(radius); dx <= int(radius); dx++ {
			x := gridX + dx
			y := gridY + dy
			if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight && wg.mask[y][x] {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= radius {
					// Impulse to velocity (not height directly)
					energy := 40.0 * (1 - dist/radius) * (1 - dist/radius)
					wg.velocity[y][x] += energy
				}
			}
		}
	}
}

func (wg *WaveGrid) update() {
	// Apply velocity to height
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			if wg.mask[y][x] {
				wg.height[y][x] += wg.velocity[y][x]
			}
		}
	}

	// Calculate new velocities using wave equation
	newVelocity := make([][]float64, gridHeight)
	for i := range newVelocity {
		newVelocity[i] = make([]float64, gridWidth)
	}

	for y := 1; y < gridHeight-1; y++ {
		for x := 1; x < gridWidth-1; x++ {
			if !wg.mask[y][x] {
				newVelocity[y][x] = 0
				continue
			}

			// Laplacian of height
			laplacian := 0.0
			neighbors := 0

			// Check 4 neighbors
			deltas := []struct{ dx, dy int }{
				{0, -1}, {0, 1}, {-1, 0}, {1, 0},
			}

			for _, d := range deltas {
				nx := x + d.dx
				ny := y + d.dy

				if nx >= 0 && nx < gridWidth && ny >= 0 && ny < gridHeight {
					if wg.mask[ny][nx] {
						laplacian += wg.height[ny][nx] - wg.height[y][x]
					} else {
						// Boundary: mirror (perfect reflection)
						laplacian += -wg.height[y][x]
					}
				}
				neighbors++
			}

			laplacian /= float64(neighbors)

			// Wave acceleration based on Laplacian
			acceleration := laplacian * waveSpeed * waveSpeed
			newVelocity[y][x] = (wg.velocity[y][x] + acceleration) * damping
		}
	}

	wg.velocity = newVelocity

	// Zero out height at boundaries
	for x := 0; x < gridWidth; x++ {
		wg.height[0][x] = 0
		wg.height[gridHeight-1][x] = 0
	}
	for y := 0; y < gridHeight; y++ {
		wg.height[y][0] = 0
		wg.height[y][gridWidth-1] = 0
	}
}

func (wg *WaveGrid) draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 25, 255})

	// Draw wave grid
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			if !wg.mask[y][x] {
				continue
			}

			h := wg.height[y][x]

			// Clamp and normalize
			h = math.Max(-80, math.Min(80, h))
			norm := h / 80.0

			var r, g, b uint8

			if norm > 0 {
				// Crest: bright blue
				b = uint8(150 + norm*100)
				g = uint8(120 + norm*60)
				r = uint8(40 + norm*40)
			} else {
				// Trough: darker, reddish
				r = uint8(100 - norm*80)
				g = uint8(100 - norm*60)
				b = uint8(120 - norm*40)
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
			vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), 2, color.RGBA{200, 150, 100, 255}, false)
		}
		// Close the shape
		p1 := wg.shape[len(wg.shape)-1]
		p2 := wg.shape[0]
		vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), 2, color.RGBA{200, 150, 100, 255}, false)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nClick to create waves", ebiten.CurrentTPS()))
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
