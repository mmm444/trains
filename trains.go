package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"github.com/ajstarks/svgo"
)

var (
	// AngleN tells how many curve parts form a full circle
	AngleN = 12
	// Radius is the radius of the circle made by curve parts
	Radius = 2.0
	// BridgeLen defines the length of the bridge part
	BridgeLen = 8.0
	// StraightLen defines the length of the straight part
	StraightLen = 1.0
)

// Angle represents the rotation of a track piece.
type Angle int

// Add adds given angle to this one. The angles are added modulo 2pi.
func (a Angle) Add(b Angle) Angle {
	res := (a + b) % Angle(AngleN)
	if res < 0 {
		res += Angle(AngleN)
	}
	return res
}

// Rad returns the angle in radians.
func (a Angle) Rad() float64 {
	return 2 * math.Pi * float64(a) / float64(AngleN)
}

//Deg returns the angle in degrees.
func (a Angle) Deg() float64 {
	return float64(a) * 360 / float64(AngleN)
}

// Kind specifies the kinds of track parts.
type Kind int

// All possible kinds of track parts.
const (
	End Kind = iota
	Bridge
	Straight
	Left
	Right
	kindCnt
)

func (k *Kind) String() string {
	return []string{"E", "B", "S", "A", "C"}[*k]
}

// Part represents a placed part in the track
type Part struct {
	X     float64 // x coord of the part origin
	Y     float64 // y coord of the pard origin
	Angle Angle   // directon the part is rotated to
	Kind  Kind    // part type
}

// AtSamePlaceAs tells whether p2 starts at the same place and has same angle as this part.
// There is some tolerance.
func (p *Part) AtSamePlaceAs(p2 *Part) bool {
	const prec = 1e-2
	return math.Abs(p.X-p2.X) < prec && math.Abs(p.Y-p2.Y) < prec && p.Angle == p2.Angle
}

type update struct {
	x, y float64
	a    Angle
}

// updateTab tells for every possible Kind and Part combination what difference
// it makes to X, Y and Angle to move to the next Part
var updateTab []update

// Update places target part after this one
func (p *Part) Update(target *Part) {
	u := &updateTab[int(p.Kind)*AngleN+int(p.Angle)]
	target.X = p.X + u.x
	target.Y = p.Y + u.y
	target.Angle = p.Angle.Add(u.a)
}

func initUpdateTab() {
	updateTab = make([]update, int(kindCnt)*AngleN)

	// curveLen is the distance between start and end of a curve
	n := float64(AngleN)
	curveLen := Radius * 2 * math.Sin(math.Pi/n)

	for a := 0; a < AngleN; a++ {
		s, c := math.Sincos(Angle(a).Rad())
		updateTab[int(Bridge)*AngleN+a] = update{
			x: c * BridgeLen,
			y: s * BridgeLen,
			a: 0,
		}
		updateTab[int(Straight)*AngleN+a] = update{
			x: c * StraightLen,
			y: s * StraightLen,
			a: 0,
		}
		sl, cl := math.Sincos(Angle(a).Rad() + math.Pi/n)
		updateTab[int(Left)*AngleN+a] = update{
			x: cl * curveLen,
			y: sl * curveLen,
			a: 1,
		}
		sr, cr := math.Sincos(Angle(a).Rad() - math.Pi/n)
		updateTab[int(Right)*AngleN+a] = update{
			x: cr * curveLen,
			y: sr * curveLen,
			a: -1,
		}
	}
}

// findTracks searches the state space of all possible pieces combinations for
// closed tracks that use all pieces. There are some symmetry branch cuts such
// as that the first piece is always the bridge if it present and that the first
// curve always goes to the left.
func findTracks(out chan []Part, bridges, straights, curves int) {
	if bridges > 1 {
		panic("can handle only 0 or 1 bridges")
	}

	track := make([]Part, bridges+straights+curves+1)
	startAt := 0
	if bridges == 1 {
		track[0].Kind = Bridge
		track[0].Update(&track[1])
		startAt = 1
	}

	n1 := len(track) - 1
	curveLen := Radius * 2 * math.Sin(math.Pi/float64(AngleN))

	var r func(int, Kind, int, int)
	r = func(pos int, kind Kind, s, c int) {
		p := &track[pos]

		// update next part start
		p.Kind = kind
		p.Update(&track[pos+1])

		// terminate search?
		if s == 0 && c == 0 {
			if track[0].AtSamePlaceAs(&track[n1]) {
				c := make([]Part, n1)
				copy(c, track[:n1])
				out <- c
			}
			return
		}

		// optimization - check that origin is still within reachable distance
		p = &track[pos+1]
		dist := math.Sqrt(p.X*p.X + p.Y*p.Y)
		if dist > float64(s)*StraightLen+float64(c)*curveLen+1 {
			// fmt.Println("Cutting on ", pos, dist)
			return
		}

		if s > 0 {
			r(pos+1, Straight, s-1, c)
		}
		if c > 0 {
			r(pos+1, Left, s, c-1)
			if c < curves { // first curve turns always to the left
				r(pos+1, Right, s, c-1)
			}
		}
	}

	// start recursion - left only - eliminate symmetrical tracks
	r(startAt, Straight, straights-1, curves)
	r(startAt, Left, straights, curves-1)

	close(out)
}

// simpleFormat formats the the track as letters
func simpleFormat(track []Part) string {
	b := bytes.Buffer{}
	for i := range track {
		b.WriteString(track[i].Kind.String())
	}
	return b.String()
}

// trackBounds finds the bounding box of the track.
func trackBounds(track []Part) (minX, minY, maxX, maxY float64) {
	for i := range track {
		p := &track[i]
		if minX > p.X {
			minX = p.X
		}
		if maxX < p.X {
			maxX = p.X
		}
		if minY > p.Y {
			minY = p.Y
		}
		if maxY < p.Y {
			maxY = p.Y
		}
	}
	return
}

//writeSvg write the track as an SVG image.
func writeSvg(writer io.Writer, track []Part) error {
	const unit = 40
	const gap = 5

	c := svg.New(writer)
	minX, minY, maxX, maxY := trackBounds(track)
	offX, offY := -minX*unit+gap, -minY*unit+gap
	w, h := int((maxX-minX)*unit+2*gap), int((maxY-minY)*unit+2*gap)
	c.Start(w, h)
	c.ScaleXY(1, -1)
	c.Translate(0, -h)

	r := int(Radius * float64(unit))
	n := float64(AngleN + 1)
	r2 := unit * Radius * 2 * math.Sin(math.Pi/n)
	x := int(r2 * math.Cos(math.Pi/n))
	y := int(r2 * math.Sin(math.Pi/n))

	for i := range track {
		p := &track[i]
		c.TranslateRotate(int(p.X*unit+offX), int(p.Y*unit+offY), p.Angle.Deg())
		switch p.Kind {
		case Straight:
			c.Line(0, 0, int(StraightLen*unit)-gap, 0, "fill:none;stroke:red;stroke-width:5")
		case Bridge:
			c.Line(0, 0, int(BridgeLen*unit)-gap, 0, "fill:none;stroke:orange;stroke-width:5")
		case Left:
			c.Arc(0, 0, r, r, 0, false, true, x, y, "fill:none;stroke:green;stroke-width:5")
		case Right:
			c.Arc(0, 0, r, r, 0, false, false, x, -y, "fill:none;stroke:blue;stroke-width:5")
		}
		c.Gend() // TranslateRotate
	}
	c.Gend() //Translate
	c.Gend() //Scale

	c.End()

	return nil
}

// writeSvgFile saves the track as an image to the SVG file with given name.
func writeSvgFile(fn string, track []Part) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeSvg(f, track)
}

// totalAngle returns the sum of angles in the track. For closed track it will
// always have the form k*AngleN where k is an integer. When it is 0 the track has
// an 8 shape. When it is AngleN the track has an O shape.
func totalAngle(track []Part) (angle int) {
	for i := range track {
		p := &track[i]
		switch p.Kind {
		case Left:
			angle++
		case Right:
			angle--
		}
	}
	return
}

// setIkeaParams adusts the part paramters for IKEA LILLABO 20 piece train set
func setIkeaParams() {
	AngleN = 8
	Radius = 1
	BridgeLen = 2
	StraightLen = 1
}

func main() {
	var (
		// lego basic set params
		brCnt   = flag.Int("b", 1, "number of bridge pieces (may 0 or 1)")
		strCnt  = flag.Int("s", 5, "number of straight pieces")
		curvCnt = flag.Int("c", 20, "number of curve pieces")

		ikea = flag.Bool("ikea", false, "toggle piece characteristics to IKEA LILLABO parts (default is LEGO Duplo)")

		only8 = flag.Bool("8", false, "output only 8 shaped tracks")
		onlyO = flag.Bool("O", false, "output only O shaped tracks")
	)

	flag.Parse()

	if *ikea {
		setIkeaParams()
	}

	initUpdateTab()

	ch := make(chan []Part)
	go findTracks(ch, *brCnt, *strCnt, *curvCnt)

	os.Mkdir("svg", 0755)

	o, err := os.Create("all.html")
	if err != nil {
		log.Fatal("cannot open output file all.html:", err)
	}
	defer o.Close()

	o.Write([]byte("<html><body>"))

	i := 0
	for tr := range ch {
		if *only8 && totalAngle(tr) != 0 || *onlyO && totalAngle(tr) == 0 {
			continue
		}
		fmt.Println(simpleFormat(tr))
		fn := fmt.Sprintf("svg/img_%05d.svg", i)
		fmt.Fprintf(o, `<img src="%s"><br>`, fn)
		writeSvgFile(fn, tr)
		i++
	}

	o.Write([]byte("</body></html>"))
}
