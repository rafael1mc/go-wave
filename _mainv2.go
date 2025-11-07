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
	gridSize     = 5
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

				// Color based on wave height
				var c color.RGBA
				if height > 0 {
					c = color.RGBA{100, 200, uint8(150 + height*100), uint8(200 + height*50)}
				} else {
					c = color.RGBA{50, uint8(100 - height*100), 150, 150}
				}

				// Draw point with size based on wave height
				radius := float32(1.0 + height*3)
				vector.DrawFilledCircle(screen, float32(x), float32(y), radius, c, false)
			}
		}
	}

	ebitenutil.DebugPrint(screen, "Click inside the circle to create waves")
}

func (g *Game) calculateWaveHeight(x, y float64) float64 {
	totalHeight := 0.0

	for _, source := range g.waveSources {
		timeSinceCreation := g.frame - source.createdAt
		waveSpeed := 2.5
		waveLength := 30.0
		amplitude := 1.0
		damping := 0.98

		dx := x - source.x
		dy := y - source.y
		distFromSource := math.Sqrt(dx*dx + dy*dy)

		// Boundary reflection
		reflectedHeight := 0.0
		for _, reflDist := range []float64{0, 0, 0} {
			bounceCount := int(reflDist)
			distToEdge := shapeRadius - distFromSource
			if distToEdge > 0 && bounceCount == 0 {
				wavePhase := float64(timeSinceCreation)*waveSpeed - (distFromSource / waveLength)
				height := amplitude * math.Sin(wavePhase) * math.Exp(-distFromSource/200)
				reflectedHeight += height
			}
		}

		// Direct wave
		wavePhase := float64(timeSinceCreation)*waveSpeed - (distFromSource / waveLength)
		height := amplitude * math.Sin(wavePhase) * math.Exp(-distFromSource/200)

		// Damping over time
		damped := height * math.Pow(damping, float64(timeSinceCreation))

		totalHeight += damped

		// Simple boundary reflection - mirror the wave at the edge
		distToEdge := shapeRadius - distFromSource
		if distToEdge > 0 && distToEdge < 100 {
			reflectionInfluence := (100 - distToEdge) / 100
			wavePhase := float64(timeSinceCreation)*waveSpeed - ((shapeRadius - distFromSource) / waveLength)
			reflectedWave := amplitude * math.Sin(wavePhase) * math.Exp(-(shapeRadius-distFromSource)/200) * reflectionInfluence
			totalHeight += reflectedWave * 0.5
		}
	}

	// Clamp height
	if totalHeight > 1.0 {
		totalHeight = 1.0
	}
	if totalHeight < -1.0 {
		totalHeight = -1.0
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
