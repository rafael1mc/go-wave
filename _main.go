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
	shapeRadius  = 150
	centerX      = screenWidth / 2
	centerY      = screenHeight / 2
)

type Particle struct {
	x, y         float64
	vx, vy       float64
	age          float64
	maxAge       float64
	distFromEdge float64
}

type Game struct {
	particles []Particle
	waves     []Wave
}

type Wave struct {
	x, y      float64
	radius    float64
	maxRadius float64
	vx, vy    float64
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		x, y := float64(mx), float64(my)

		// Check if click is inside the shape
		dx := x - centerX
		dy := y - centerY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < shapeRadius {
			// Create wave at click position
			g.waves = append(g.waves, Wave{
				x:         x,
				y:         y,
				radius:    5,
				maxRadius: 300,
				vx:        0,
				vy:        0,
			})
		}
	}

	// Update waves
	for i := 0; i < len(g.waves); i++ {
		g.waves[i].radius += 2.5

		// Remove waves that are too large
		if g.waves[i].radius > g.waves[i].maxRadius {
			g.waves = append(g.waves[:i], g.waves[i+1:]...)
			i--
		}
	}

	// Generate particles from waves
	for _, w := range g.waves {
		// Create particles along the wave front
		numParticles := int(w.radius / 5)
		if numParticles > 2 {
			numParticles = 2
		}

		for j := 0; j < numParticles; j++ {
			angle := float64(j) * 2 * math.Pi / float64(numParticles)

			px := w.x + w.radius*math.Cos(angle)
			py := w.y + w.radius*math.Sin(angle)

			// Check if particle is within shape
			dx := px - centerX
			dy := py - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < shapeRadius {
				// Calculate outward normal
				nx := dx / dist
				ny := dy / dist

				p := Particle{
					x:      px,
					y:      py,
					vx:     nx * 2,
					vy:     ny * 2,
					age:    0,
					maxAge: 0.8,
				}
				g.particles = append(g.particles, p)
			}
		}
	}

	// Update particles
	for i := 0; i < len(g.particles); i++ {
		p := &g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.age += 0.016

		// Check boundary collision
		dx := p.x - centerX
		dy := p.y - centerY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > shapeRadius {
			// Calculate normal at boundary
			nx := dx / dist
			ny := dy / dist

			// Reflect velocity
			dotProduct := p.vx*nx + p.vy*ny
			p.vx = (p.vx - 2*dotProduct*nx) * 0.95
			p.vy = (p.vy - 2*dotProduct*ny) * 0.95

			// Push back inside
			p.x = centerX + nx*(shapeRadius-2)
			p.y = centerY + ny*(shapeRadius-2)
		}

		// Apply damping
		p.vx *= 0.98
		p.vy *= 0.98

		// Remove dead particles
		if p.age > p.maxAge {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
			i--
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 20, 30, 255})

	// Draw boundary circle
	vector.StrokePath(screen, &vector.Path{
		drawCircle(centerX, centerY, shapeRadius, 128), 2, color.RGBA{100, 150, 200, 255}, false,
	},
	)

	// Draw waves
	for _, w := range g.waves {
		alpha := uint8(200 * (1 - w.radius/w.maxRadius))
		vector.StrokePath(screen, drawCircle(w.x, w.y, w.radius, 64), 1.5, color.RGBA{100, 200, 255, alpha}, false)
	}

	// Draw particles
	for _, p := range g.particles {
		alpha := uint8(255 * (1 - p.age/p.maxAge))
		c := color.RGBA{150, 220, 255, alpha}
		vector.DrawFilledRect(screen, float32(p.x)-1, float32(p.y)-1, 2, 2, c)
	}

	// Draw instructions
	ebitenutil.DebugPrint(screen, "Click inside the circle to create waves")
	// ebitenutil.DebugPrintString(screen, "Click inside the circle to create waves")
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func drawCircle(cx, cy, r float64, segments int) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, segments)
	for i := 0; i < segments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(segments)
		x := cx + r*math.Cos(angle)
		y := cy + r*math.Sin(angle)
		vertices[i].DstX = float32(x)
		vertices[i].DstY = float32(y)
		vertices[i].ColorR = 1
		vertices[i].ColorG = 1
		vertices[i].ColorB = 1
		vertices[i].ColorA = 1
	}
	return vertices
}

func main() {
	g := &Game{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Wave Propagation Simulator")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
