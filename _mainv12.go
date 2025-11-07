package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1000
	screenHeight = 800
	shapeRadius  = 200
	centerX      = screenWidth / 2
	centerY      = screenHeight / 2
	gridSize     = 4
)

type WaveSource struct {
	x, y      float64
	createdAt int
}

type Game struct {
	waveSources []WaveSource
	pressed     bool
	frame       int
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !g.pressed {
			mx, my := ebiten.CursorPosition()
			x := float64(mx)
			y := float64(my)

			dx := x - float64(centerX)
			dy := y - float64(centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < float64(shapeRadius) {
				g.waveSources = append(g.waveSources, WaveSource{x, y, g.frame})
			}
			g.pressed = true
		}
	} else {
		g.pressed = false
	}

	g.frame++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 20, 30, 255})

	// Draw boundary circle
	vector.StrokeCircle(screen, float32(centerX), float32(centerY), float32(shapeRadius), 2, color.RGBA{100, 150, 200, 255}, false)

	// Draw grid of points
	for xi := int(centerX - float64(shapeRadius)); xi < int(centerX+float64(shapeRadius)); xi += gridSize {
		for yi := int(centerY - float64(shapeRadius)); yi < int(centerY+float64(shapeRadius)); yi += gridSize {
			px := float64(xi)
			py := float64(yi)

			dx := px - float64(centerX)
			dy := py - float64(centerY)
			distFromCenter := math.Sqrt(dx*dx + dy*dy)

			// Only draw if inside circle
			if distFromCenter < float64(shapeRadius) {
				height := g.calculateWaveHeight(px, py)

				// Map height to color - simple and clean
				var r, g_val, b uint8
				r = 100
				g_val = 150
				b = 200

				if height > 0.1 {
					// Positive: brighter blue
					r = 80
					g_val = 180
					b = 255
				} else if height < -0.1 {
					// Negative: orange/red
					r = 255
					g_val = 120
					b = 80
				}

				c := color.RGBA{r, g_val, b, 255}

				// Draw point with size based on wave height
				radius := float32(1.0 + math.Abs(height)*3)
				vector.DrawFilledCircle(screen, float32(px), float32(py), radius, c, false)
			}
		}
	}

	ebitenutil.DebugPrintString(screen, "Click inside the circle to create waves")
}

func (g *Game) calculateWaveHeight(x, y float64) float64 {
	totalHeight := 0.0

	for _, source := range g.waveSources {
		// Direct wave from source
		dx := x - source.x
		dy := y - source.y
		distFromSource := math.Sqrt(dx*dx + dy*dy)

		waveSpeed := 1.5
		wavelength := 40.0
		amplitude := 1.0
		timeElapsed := float64(g.frame - source.createdAt)
		waveFront := waveSpeed * timeElapsed

		// Outgoing wave
		if distFromSource < waveFront {
			distanceFromFront := distFromSource - waveFront
			waveInfluence := 30.0

			if math.Abs(distanceFromFront) < waveInfluence {
				envelope := math.Exp(-(distanceFromFront * distanceFromFront) / (waveInfluence * waveInfluence))
				phase := (distFromSource / wavelength) * 2 * math.Pi
				wave := amplitude * math.Sin(phase) * envelope
				totalHeight += wave
			}
		}

		// Reflected waves from boundary
		// Sample points on the boundary for reflection
		for angle := 0.0; angle < 2*math.Pi; angle += 0.1 {
			boundaryX := float64(centerX) + float64(shapeRadius)*math.Cos(angle)
			boundaryY := float64(centerY) + float64(shapeRadius)*math.Sin(angle)

			// Distance from source to this boundary point
			distSourceToBoundary := math.Sqrt((boundaryX-source.x)*(boundaryX-source.x) + (boundaryY-source.y)*(boundaryY-source.y))

			// Time when wave reaches this boundary point
			timeToReachBoundary := distSourceToBoundary / waveSpeed

			if timeElapsed > timeToReachBoundary {
				// Time since reflection at this point
				timeSinceReflection := timeElapsed - timeToReachBoundary

				// Reflected wave front distance (travels inward from boundary)
				reflectedWaveFront := waveSpeed * timeSinceReflection

				// Distance from current point to reflection point
				reflDx := x - boundaryX
				reflDy := y - boundaryY
				distFromReflection := math.Sqrt(reflDx*reflDx + reflDy*reflDy)

				if distFromReflection < reflectedWaveFront {
					distFromReflectedFront := distFromReflection - reflectedWaveFront
					waveInfluence := 30.0

					if math.Abs(distFromReflectedFront) < waveInfluence {
						envelope := math.Exp(-(distFromReflectedFront * distFromReflectedFront) / (waveInfluence * waveInfluence))
						phase := (distFromReflection / wavelength) * 2 * math.Pi
						wave := amplitude * math.Sin(phase) * envelope * 0.6
						totalHeight += wave
					}
				}
			}
		}
	}

	// Clamp
	if totalHeight > 1.0 {
		return 1.0
	}
	if totalHeight < -1.0 {
		return -1.0
	}
	return totalHeight
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Wave Propagation Simulator")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
