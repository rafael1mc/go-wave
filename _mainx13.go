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
	waveDamping  = 0.995
	waveSpeed    = 2.0
)

type Particle struct {
	x, y    float64
	vx, vy  float64
	origX   float64
	origY   float64
	onShape bool
}

type Wave struct {
	particles []*Particle
	shape     []Vector2
}

type Vector2 struct {
	x, y float64
}

func NewWave() *Wave {
	w := &Wave{
		particles: make([]*Particle, 0),
		shape:     generateCircleShape(screenWidth/2, screenHeight/2, 150),
	}
	w.initializeParticles()
	return w
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

func (w *Wave) initializeParticles() {
	for _, point := range w.shape {
		p := &Particle{
			x:       point.x,
			y:       point.y,
			origX:   point.x,
			origY:   point.y,
			vx:      0,
			vy:      0,
			onShape: true,
		}
		w.particles = append(w.particles, p)
	}
}

func (w *Wave) addWaveAtMouse(mx, my float64) {
	// Find closest particle to mouse click
	minDist := math.MaxFloat64
	var closestP *Particle
	for _, p := range w.particles {
		dist := math.Sqrt((p.x-mx)*(p.x-mx) + (p.y-my)*(p.y-my))
		if dist < minDist {
			minDist = dist
			closestP = p
		}
	}

	if closestP != nil && minDist < 100 {
		// Apply impulse to create wave
		closestP.vy -= 15
	}
}

func (w *Wave) update() {
	// Update particle velocities and positions
	for i, p := range w.particles {
		if !p.onShape {
			continue
		}

		// Apply velocity
		p.x += p.vx
		p.y += p.vy

		// Spring force back to original position
		dx := p.origX - p.x
		dy := p.origY - p.y
		springForce := 0.05
		p.vx += dx * springForce
		p.vy += dy * springForce

		// Damping
		p.vx *= waveDamping
		p.vy *= waveDamping

		// Propagate wave to neighbors
		// neighborDist := 15.0
		leftIdx := (i - 1 + len(w.particles)) % len(w.particles)
		rightIdx := (i + 1) % len(w.particles)

		leftP := w.particles[leftIdx]
		rightP := w.particles[rightIdx]

		// Wave propagation
		if math.Abs(p.y-p.origY) < 100 { // Only propagate if not too large
			spread := 0.2
			p.vy += spread * (leftP.y - p.y)
			p.vy += spread * (rightP.y - p.y)
		}
	}
}

func (w *Wave) draw(screen *ebiten.Image) {
	screen.Fill(color.Color(color.RGBA{20, 20, 30, 255}))

	// Draw the shape
	if len(w.shape) > 1 {
		for i := 0; i < len(w.particles)-1; i++ {
			p1 := w.particles[i]
			p2 := w.particles[i+1]
			vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), float32(2), color.RGBA{100, 200, 255, 255}, false)
		}
		// Close the shape
		p1 := w.particles[len(w.particles)-1]
		p2 := w.particles[0]
		vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), float32(2), color.RGBA{100, 200, 255, 255}, false)
	}

	// Draw particles
	for _, p := range w.particles {
		offsetFromOriginal := math.Abs(p.y - p.origY)
		intensity := uint8(math.Min(255, offsetFromOriginal*2))
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), float32(3), color.RGBA{100 + intensity, 150, 255, 255}, false)
	}
}

type Game struct {
	wave *Wave
}

func NewGame() *Game {
	return &Game{
		wave: NewWave(),
	}
}

func (g *Game) Update() error {
	// Mouse click detection
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.wave.addWaveAtMouse(float64(x), float64(y))
	}

	g.wave.update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.wave.draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nClick on the shape to create waves", ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Wave Simulation")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
