package planetarium

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
)

type Vector struct {
	X, Y float32
}

func (v Vector) String() string {
	return fmt.Sprintf(
		"[%6.1fx, %6.1fy] (rad: %6.4f, deg: %6.2f) (rad: %6.4f, deg: %6.2f)",
		v.X,
		v.Y,
		v.Normalize(),
		v.Normalize().Degree(),
		FromDirection(v.Normalize(), 5).Normalize(),
		FromDirection(v.Normalize(), 5).Normalize().Degree())
}

func FromDirection(r Radian, length float32) Vector {
	y := float64(length) * math.Sin(float64(r))
	x := math.Sqrt(float64(length)*float64(length) - y*y)

	return Vector{float32(x), float32(y)}
}

// returns global angle of vector in radians
func (v Vector) Normalize() Radian { // todo fix
	x, y := float64(v.X), float64(v.Y)
	c := math.Sqrt(x*x + y*y)
	return Radian(math.Asin(y / c))

}

type Radian float32

// converts to standard degrees
func (r Radian) Degree() float32 {
	return float32(r) * 360 / (2 * math.Pi)
}

type Position struct {
	X, Y float32
}

func (p Position) String() string {
	return fmt.Sprintf("[%6.1fx, %6.1fy]", p.X, p.Y)
}

type Planet struct {
	Pos    Position
	Vector Vector

	Mass   float32 //mass
	Radius float32 //radius
}

func NewPlanet(x, y, r float32) *Planet {
	pos := Position{x, y}

	return &Planet{Pos: pos, Radius: r}
}

// Calculates relative distance between given Planet
func (p *Planet) Distance(p2 *Planet) float32 {
	dx, dy := p.Pos.X-p2.Pos.X, p.Pos.Y-p2.Pos.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	return float32(math.Sqrt(math.Pow(float64(dx), 2) + math.Pow(float64(dy), 2)))
}

func (p *Planet) Area() float64 {
	return math.Pow(float64(p.Radius), 2) * math.Pi
}

// modify internal Vector according to met Planet
func (p *Planet) Bounce(p2 *Planet) {
	p.Vector.X = -p.Vector.X
	p.Vector.Y = -p.Vector.Y
}

// merge other Planet
func (p *Planet) Merge(p2 *Planet) {
	newRadius := math.Sqrt((p.Area() + p2.Area()) / math.Pi)

	p.Radius = float32(newRadius)
	p.Mass += p2.Mass

	// decreasing velocity todo: make it proper, move to "Add partially"
	p.Vector.X -= p.Vector.X / 20
	p.Vector.Y -= p.Vector.Y / 20

}

var TooSmall = errors.New("cant partially merge planets")

// partially merge Planet
func (p *Planet) AddPartialyl(p2 *Planet) error {
	distance := float64(p.Distance(p2))
	if distance < 0 {
		distance = -distance
	}

	overlap := float64(p.Radius+p2.Radius) - distance
	if overlap < 0 {
		overlap = -overlap
	}

	if float64(p2.Radius) < overlap {
		//p.Merge(p2) // can't be handled there
		//p2.Radius = 0 // todo: fix
		return TooSmall
	}

	// thanks to zskk for finding algorithm
	g := overlap*overlap/2 - float64(p2.Radius)*overlap
	x1 := (-distance + math.Sqrt(math.Pow(distance, 2)-(4*g))) / 2
	x2 := (-distance - math.Sqrt(math.Pow(distance, 2)-(4*g))) / 2

	var x float64

	switch {
	case x1 > 0:
		x = x1
	case x2 > 0:
		x = x2
	default:
		x = 0
	}

	p.Radius = p.Radius + float32(x)
	p2.Radius = p2.Radius + float32(x-overlap)

	return nil
}

func GetRandomPlanets(min, max int, maxx, maxy int32) (planets []*Planet) {
	var random int
	if min == max {
		random = 0
	} else {
		random = rand.Intn(max - min)
	}

	scope := random + min

	for i := 0; i <= scope; i++ {
		x, y := float32(rand.Int31()%maxx), float32(rand.Int31()%maxy)
		vx := (rand.Float32()*2 - 1.0) * 10
		vy := (rand.Float32()*2 - 1.0) * 10
		r := 3 + rand.Float32()*20
		m := rand.Float32() * 10

		p := Planet{
			Position{x, y},
			Vector{vx, vy},
			m,
			r,
		}

		planets = append(planets, &p)
	}

	return
}
