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

	// Sum contributions from all wave sources
	// This creates superposition - waves add together
	for _, source := range g.waveSources {
		// Calculate outgoing wave height at this point
		totalHeight += g.calculateOutgoingWave(x, y, source)

		// Calculate reflected waves from all boundary points this source hits
		totalHeight += g.calculateReflectedWaves(x, y, source)
	}

	// Clamp height - this creates interference patterns
	// When waves add constructively, amplitude increases
	// When waves add destructively, they cancel out
	if totalHeight > 1.5 {
		return 1.5
	}
	if totalHeight < -1.5 {
		return -1.5
	}
	return totalHeight
}

func (g *Game) calculateOutgoingWave(x, y float64, source WaveSource) float64 {
	dx := x - source.x
	dy := y - source.y
	distFromSource := math.Sqrt(dx*dx + dy*dy)

	waveSpeed := 1.5
	wavelength := 40.0
	amplitude := 1.5
	timeElapsed := float64(g.frame - source.createdAt)
	waveFront := waveSpeed * timeElapsed

	if distFromSource < waveFront {
		distanceFromFront := distFromSource - waveFront
		waveInfluence := 30.0

		if math.Abs(distanceFromFront) < waveInfluence {
			envelope := math.Exp(-(distanceFromFront * distanceFromFront) / (waveInfluence * waveInfluence))
			phase := (distFromSource / wavelength) * 2 * math.Pi
			wave := amplitude * math.Sin(phase) * envelope
			damping := math.Exp(-timeElapsed / 300)
			return wave * damping
		}
	}
	return 0.0
}

func (g *Game) calculateReflectedWaves(x, y float64, source WaveSource) float64 {
	waveSpeed := 1.5
	wavelength := 40.0
	amplitude := 1.5
	timeElapsed := float64(g.frame - source.createdAt)

	// Only calculate reflections if enough time has passed
	if timeElapsed < 50 {
		return 0.0
	}

	totalReflectedHeight := 0.0

	// Instead of sampling all boundary points, calculate the reflection analytically
	// Find the angle from center to the current point
	pointDx := x - float64(centerX)
	pointDy := y - float64(centerY)
	pointAngle := math.Atan2(pointDy, pointDx)

	// For the reflection, we need to find where on the boundary the wave will reflect
	// The wave travels from source outward, so it hits different boundary points at different times
	// We'll sample a few angles around the current point's angle for efficiency

	sampleAngles := 8
	for i := 0; i < sampleAngles; i++ {
		// Sample around the point's angle
		offsetAngle := float64(i-sampleAngles/2) * (math.Pi / float64(sampleAngles))
		angle := pointAngle + offsetAngle

		// Point on the boundary
		boundaryX := float64(centerX) + float64(shapeRadius)*math.Cos(angle)
		boundaryY := float64(centerY) + float64(shapeRadius)*math.Sin(angle)

		// Distance from source to this boundary point
		distToBoundary := math.Sqrt((boundaryX-source.x)*(boundaryX-source.x) + (boundaryY-source.y)*(boundaryY-source.y))

		// Time when the wave from source reaches this boundary point
		timeToReachBoundary := distToBoundary / waveSpeed

		// Has the wave reached this boundary point yet?
		if timeElapsed > timeToReachBoundary {
			timeSinceReflection := timeElapsed - timeToReachBoundary

			// Distance the reflected wave has traveled from this boundary point
			reflectedWaveFront := waveSpeed * timeSinceReflection

			// Distance from current point to this boundary reflection point
			reflDx := x - boundaryX
			reflDy := y - boundaryY
			distFromReflectionPoint := math.Sqrt(reflDx*reflDx + reflDy*reflDy)

			// Check if reflected wave reaches this point
			if distFromReflectionPoint < reflectedWaveFront {
				distanceFromReflectedFront := distFromReflectionPoint - reflectedWaveFront
				waveInfluence := 30.0

				if math.Abs(distanceFromReflectedFront) < waveInfluence {
					envelope := math.Exp(-(distanceFromReflectedFront * distanceFromReflectedFront) / (waveInfluence * waveInfluence))
					phase := (distFromReflectionPoint / wavelength) * 2 * math.Pi
					wave := amplitude * math.Sin(phase) * envelope * 0.7
					damping := math.Exp(-timeElapsed / 350)
					totalReflectedHeight += wave * damping
				}
			}
		}
	}

	return totalReflectedHeight / float64(sampleAngles)
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
