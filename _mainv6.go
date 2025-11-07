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

				// Map height to color
				var r, g_val, b uint8
				if height > 0 {
					// Positive: light blue
					intensity := uint8(math.Min(height*200, 255))
					r = 100
					g_val = uint8(150 + int(intensity)/2)
					b = 220
				} else {
					// Negative: light orange/red
					intensity := uint8(math.Min(-height*200, 255))
					r = 220
					g_val = uint8(150 - int(intensity)/2)
					b = 100
				}

				c := color.RGBA{r, g_val, b, 200}

				// Draw point with size based on wave height
				radius := float32(math.Max(1.0, 1.5+math.Abs(height)*2))
				vector.DrawFilledCircle(screen, float32(px), float32(py), radius, c, false)
			}
		}
	}

	ebitenutil.DebugPrint(screen, "Click inside the circle to create waves")
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

		// Distance from the wave front
		distanceFromFront := distFromSource - waveFront

		// Only create waves after they've started propagating
		if distFromSource < waveFront {
			waveInfluence := 30.0
			if math.Abs(distanceFromFront) < waveInfluence {
				// Gaussian envelope
				envelope := math.Exp(-(distanceFromFront * distanceFromFront) / (waveInfluence * waveInfluence))

				// Sinusoidal wave
				phase := (distFromSource / wavelength) * 2 * math.Pi
				wave := amplitude * math.Sin(phase) * envelope

				// Damping over time
				damping := math.Exp(-timeElapsed / 300)
				height := wave * damping

				totalHeight += height
			}
		}

		// Handle boundary reflection
		distToEdge := float64(shapeRadius) - distFromSource
		if distToEdge > 0 && distToEdge < 150 {
			reflectedWaveFront := waveSpeed*timeElapsed - 2*(float64(shapeRadius)-distFromSource)
			reflectedDistanceFromFront := distFromSource - reflectedWaveFront

			if math.Abs(reflectedDistanceFromFront) < 30 && reflectedWaveFront > 0 {
				envelope := math.Exp(-(reflectedDistanceFromFront * reflectedDistanceFromFront) / (30 * 30))
				phase := (distFromSource / wavelength) * 2 * math.Pi
				wave := amplitude * math.Sin(phase) * envelope * 0.6
				damping := math.Exp(-timeElapsed / 300)
				height := wave * damping
				totalHeight += height
			}
		}
	}

	// Clamp height
	if totalHeight > 1.5 {
		return 1.5
	}
	if totalHeight < -1.5 {
		return -1.5
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
