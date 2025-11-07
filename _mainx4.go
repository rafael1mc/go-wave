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
			x, y := float64(mx), float64(my)

			dx := x - centerX
			dy := y - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < shapeRadius {
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
	for x := centerX - shapeRadius; x < centerX+shapeRadius; x += gridSize {
		for y := centerY - shapeRadius; y < centerY+shapeRadius; y += gridSize {
			dx := x - centerX
			dy := y - centerY
			distFromCenter := math.Sqrt(float64(dx)*float64(dx) + float64(dy)*float64(dy))

			// Only draw if inside circle
			if distFromCenter < shapeRadius {
				height := g.calculateWaveHeight(float64(x), float64(y))

				// Map height to color - smooth gradient
				normalizedHeight := height / 1.5
				if normalizedHeight > 0 {
					// Positive: light blue
					intensity := uint8(normalizedHeight * 200)
					r = 100
					g_c = 150 + intensity/2
					b = 220
				} else {
					// Negative: light orange/red
					intensity := uint8(-normalizedHeight * 200)
					r = 220
					g_c = 150 - intensity/2
					b = 100
				}

				c := color.RGBA{r, g_c, b, 200}

				// Draw point with size based on wave height
				radius := float32(math.Max(1.0, 1.5+math.Abs(height)*2))
				vector.DrawFilledCircle(screen, float32(x), float32(y), radius, c, false)
			}
		}
	}

	ebitenutil.DebugPrintString(screen, "Click inside the circle to create waves")
}

func (g *Game) calculateWaveHeight(x, y float64) float64 {
	totalHeight := 0.0

	for _, source := range g.waveSources {
		dx := x - source.x
		dy := y - source.y
		distFromSource := math.Sqrt(dx*dx + dy*dy)

		// Wave parameters
		waveSpeed := 1.5
		wavelength := 40.0
		amplitude := 1.5

		// Time elapsed since wave source created
		timeElapsed := float64(g.frame - source.createdAt)

		// Position of the wave front
		waveFront := waveSpeed * timeElapsed

		// Distance from the wave front (how far ahead or behind the wave crest)
		distanceFromFront := distFromSource - waveFront

		// Only create waves after they've started propagating
		if distFromSource < waveFront {
			// Create wave oscillation
			// Only oscillate near the wave front
			waveInfluence := 30.0
			if math.Abs(distanceFromFront) < waveInfluence {
				// Gaussian envelope to smooth the wave
				envelope := math.Exp(-(distanceFromFront * distanceFromFront) / (waveInfluence * waveInfluence))

				// Sinusoidal wave
				phase := (distFromSource / wavelength) * 2 * math.Pi
				wave := amplitude * math.Sin(phase) * envelope

				// Damping over time (energy dissipation)
				damping := math.Exp(-timeElapsed / 300)
				height := wave * damping

				totalHeight += height
			}
		}

		// Handle boundary reflection
		distToEdge := shapeRadius - distFromSource
		if distToEdge > 0 && distToEdge < 150 {
			// Reflected wave comes back from the boundary
			reflectedWaveFront := waveSpeed*timeElapsed - 2*(shapeRadius-distFromSource)
			reflectedDistanceFromFront := distFromSource - reflectedWaveFront

			if math.Abs(reflectedDistanceFromFront) < waveInfluence && reflectedWaveFront > 0 {
				envelope := math.Exp(-(reflectedDistanceFromFront * reflectedDistanceFromFront) / (waveInfluence * waveInfluence))
				phase := (distFromSource / wavelength) * 2 * math.Pi
				wave := amplitude * math.Sin(phase) * envelope * 0.6
				damping := math.Exp(-timeElapsed / 300)
				height := wave * damping
				totalHeight += height
			}
		}
	}

	// Clamp height
	return math.Max(math.Min(totalHeight, 1.5), -1.5)
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
