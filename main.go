// author: Jacky Boen

package main

import (
	"./pkg/planetarium"
	"fmt"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"
)

var ( // settings
	debug            = false
	simSpeed float32 = 0.01

	// recommended: todo: handle this by randomizer itself
	// 2560x1440 - 3000, 3000
	// 1920x1080 - 1500, 1500
	minPlanets, maxPlanets = 1500, 1500

	playerRadius float32 = 10.0
	winX, winY   int32   = 1920, 1080
)

var (
	smallerPlanetColor       = sdl.Color{127, 127, 255, 255}
	almostSmallerPlanetColor = sdl.Color{255, 127, 30, 255}
	biggerPlanetColor        = sdl.Color{255, 30, 30, 255}
	playerColor              = sdl.Color{20, 255, 40, 255}
	vectorColor              = sdl.Color{255, 0, 0, 255}
	textColor                = sdl.Color{255, 255, 255, 196}
	textWarnColor            = sdl.Color{255, 0, 0, 196}
)

var planets []*planetarium.Planet
var playerPlanet *planetarium.Planet
var speed float32
var mutex = sync.Mutex{} // sync between modifying `planets` and rendering them

var renderer *sdl.Renderer
var mouseX, mouseY int32
var helpEnabled bool

// render single planet
func renderPlanet(p *planetarium.Planet, debug bool) {
	x, y, r := int32(p.Pos.X), int32(p.Pos.Y), int32(p.Radius)

	var color sdl.Color

	switch {
	case p == playerPlanet:
		color = playerColor
	case p.Area() < playerPlanet.Area(): // todo: use mass instead of areae
		ratio := p.Area() / playerPlanet.Area()
		if ratio > 0.5 {
			treshold := (ratio-0.5)*2*-1 + 1 // 0.0 - 1.0

			almostSmallerPlanetColor.R = 255 - uint8(128*treshold)
			almostSmallerPlanetColor.B = 30 + uint8(225*treshold)
			color = almostSmallerPlanetColor
		} else {
			color = smallerPlanetColor
		}
	default:
		color = biggerPlanetColor
	}

	gfx.FilledCircleColor(renderer, x, y, r, color)

	if !debug || p != playerPlanet {
		return // skip debugging part
	}

	// render Vector
	vx, vy := int32(p.Vector.X), int32(p.Vector.Y)
	gfx.LineColor(renderer, x, y, x+vx, y+vy, vectorColor)

	// render debug data
	tXOffset := x + 10 + int32(p.Radius)

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	renderer.SetDrawColor(0, 0, 40, 127)
	renderer.FillRect(&sdl.Rect{tXOffset - 5, y - 25, 440, 50})

	gfx.StringColor(renderer, tXOffset, y-20, fmt.Sprintf("   pos: %s", p.Pos), textColor)
	gfx.StringColor(renderer, tXOffset, y-10, fmt.Sprintf("Vector: %s", p.Vector), textColor)
	gfx.StringColor(renderer, tXOffset, y, fmt.Sprintf("  mass: [%6.1f]", p.Mass), textColor)
	gfx.StringColor(renderer, tXOffset, y+10, fmt.Sprintf("radius: [%6.1f]", p.Radius), textColor)
}

// render all planets
func renderPlanets(debug bool) {
	mutex.Lock()
	var playerAlive bool
	for _, p := range planets {
		if p == playerPlanet {
			playerAlive = true
			break
		}
	}
	if playerAlive {
		renderer.SetDrawColor(0, 0, 0, 0)
	} else {
		renderer.SetDrawColor(70, 0, 0, 0)
	}
	renderer.Clear()

	for _, p := range planets {
		renderPlanet(p, debug)
	}
	mutex.Unlock()

	if !playerAlive {
		renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
		renderer.SetDrawColor(0, 0, 0, 200)
		renderer.FillRect(&sdl.Rect{winX/2 - 5, winY/2 - 5, 74, 20})
		gfx.StringColor(renderer, winX/2, winY/2, "GameOver", textWarnColor)
	} else {
		vector := planetarium.Vector{playerPlanet.Pos.X - float32(mouseX), playerPlanet.Pos.Y - float32(mouseY)}
		direction := vector.Normalize()
		newVector := planetarium.FromDirection(direction, 50.0)

		gfx.LineColor(renderer, int32(playerPlanet.Pos.X), int32(playerPlanet.Pos.Y), int32(playerPlanet.Pos.X+newVector.X), int32(playerPlanet.Pos.Y+newVector.Y), sdl.Color{200, 200, 255, 127})
	}

	if helpEnabled {
		renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
		renderer.SetDrawColor(0, 0, 0, 200)
		renderer.FillRect(&sdl.Rect{0, 0, 350, 90})

		gfx.StringColor(renderer, 5, 5, fmt.Sprintf("game speed: %5.1f%%", speed/simSpeed*100), textColor)
		gfx.StringColor(renderer, 5, 35, "(space): new game", textColor)
		gfx.StringColor(renderer, 5, 45, "(d): player's debug info", textColor)
		gfx.StringColor(renderer, 5, 55, "(lmb): thrust", textColor)
		gfx.StringColor(renderer, 5, 65, "(mmb): reset simulation speed", textColor)
		gfx.StringColor(renderer, 5, 75, "(mouse wheel): modify simulation speed", textColor)
	}

	renderer.Present()

}

func newGame() {
	speed = simSpeed
	planets = planetarium.GetRandomPlanets(minPlanets, maxPlanets, winX, winY)

	playerPlanet = planetarium.NewPlanet(float32(winX/2), float32(winY/2), playerRadius)
	planets = append(planets, playerPlanet)
}

func run() int {
	rand.Seed(time.Now().UTC().UnixNano())

	window, err := sdl.CreateWindow("xd", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winX, winY, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	//mode, _ := sdl.GetDisplayMode(0, 0)
	//winX, winY = mode.W, mode.H
	//
	//window.SetDisplayMode(&mode)
	//window.SetFullscreen(1)

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	renderer.SetDrawColor(0, 0, 0, 0)
	renderer.Clear()

	newGame()

	go func() { /// updating planets, todo: something about it
		for {

			var inCollision [][2]*planetarium.Planet
			var toRemove []*planetarium.Planet

			// fix vectors
			for _, p1 := range planets {
				//if p1.pos.mouseX < 0 || p1.pos.mouseX > float32(winX) || p1.pos.mouseY < 0 || p1.pos.mouseY > float32(winY) {
				//	toRemove = append(toRemove, p1)
				//	continue
				for _, p2 := range planets {
					if p1 == p2 {
						continue
					}

					// hope-optimization, can be removed
					a1, b1, c1, d1 := p1.Pos.X-p1.Radius, p1.Pos.X+p1.Radius, p1.Pos.Y-p1.Radius, p1.Pos.Y+p1.Radius
					a2, b2, c2, d2 := p2.Pos.X-p2.Radius, p2.Pos.X+p2.Radius, p2.Pos.Y-p2.Radius, p2.Pos.Y+p2.Radius

					if b1 < a2 || a1 > b2 || d1 < c2 || c1 > d2 {
						continue // optimization I hope
					}
					// end of hope-optimization

					treshold := p1.Radius + p2.Radius
					distance := p1.Distance(p2)

					if distance <= treshold {
						match := [2]*planetarium.Planet{p1, p2}
						inCollision = append(inCollision, match)
					}

				}

			}

		collision: // todo: learn to how to get proper name for labels
			for _, icp := range inCollision {
				p1, p2 := icp[0], icp[1]

				for _, ptr := range toRemove {
					if p1 == ptr || p2 == ptr {
						continue collision // one of Planet already marked to delete, cannot make a second merge
					}
				}

				if p1.Area() > p2.Area() { // todo: mass should make the decision, not area
					if p2.Area() < 0.1 {
						p1.Merge(p2)
						toRemove = append(toRemove, p2)
					} else {
						err := p1.AddPartialyl(p2)
						if err != nil && err == planetarium.TooSmall {
							p1.Merge(p2)
							toRemove = append(toRemove, p2)
						}
					}

				} else {
					if p1.Area() < 0.1 { // todo: damn, nice repetition
						p2.Merge(p1)
						toRemove = append(toRemove, p1)
					} else {
						err := p2.AddPartialyl(p1)
						if err != nil && err == planetarium.TooSmall {
							p2.Merge(p1)
							toRemove = append(toRemove, p1)
						}
					}
				}
			}

			var indexToRemove []int

			for _, p := range toRemove {
				for i, rp := range planets {
					if p == rp {
						indexToRemove = append(indexToRemove, i)
					}
				}
			}

			sort.Sort(sort.Reverse(sort.IntSlice(indexToRemove)))

			mutex.Lock()
			for _, v := range indexToRemove {
				planets = append(planets[:v], planets[v+1:]...)
			}
			mutex.Unlock()

			// move planets around according to its vectors
			for _, p := range planets {
				newX := float64(p.Pos.X + p.Vector.X*speed)
				newY := float64(p.Pos.Y + p.Vector.Y*speed)

				newX = math.Mod(newX, float64(winX))
				newY = math.Mod(newY, float64(winY))
				if newX < 0 {
					newX = float64(winX) - newX
				}
				if newY < 0 {
					newY = float64(winY) - newY
				}

				p.Pos.X = float32(newX)
				p.Pos.Y = float32(newY)

			}
			time.Sleep(time.Millisecond * 2)
		}
	}()

	var running bool
	var event sdl.Event

	running = true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				e := event.(*sdl.MouseButtonEvent)

				if e.Type == sdl.MOUSEBUTTONDOWN {
					switch e.Button {
					case sdl.BUTTON_LEFT:
						modifyPlayerVector(e.X, e.Y)
					case sdl.BUTTON_MIDDLE:
						speed = simSpeed
					}

				}
			case *sdl.MouseWheelEvent:
				e := event.(*sdl.MouseWheelEvent)

				speed += speed / 10 * float32(e.Y)
				if speed < 0.001 {
					speed = 0.001
				} else if speed > 10.0 {
					speed = 10.0
				}

			case *sdl.MouseMotionEvent:
				e := event.(*sdl.MouseMotionEvent)
				mouseX, mouseY = e.X, e.Y
			case *sdl.KeyboardEvent:
				e := event.(*sdl.KeyboardEvent)
				if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {
					case sdl.K_SPACE:
						newGame()
					case sdl.K_d:
						debug = !debug
					case sdl.K_h:
						helpEnabled = !helpEnabled
					}

				}
			}
		}
		renderPlanets(debug)
		sdl.Delay(1)
		window.GLSwap()
	}

	return 0
}

func modifyPlayerVector(x int32, y int32) {
	px, py := int32(playerPlanet.Pos.X), int32(playerPlanet.Pos.Y)

	dx, dy := x-px, y-py

	playerPlanet.Vector.X -= float32(dx) / 50
	playerPlanet.Vector.Y -= float32(dy) / 50
}

func main() {
	os.Exit(run())
}
