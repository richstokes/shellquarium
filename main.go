package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ANSI escape codes
const (
	ESC      = "\033["
	ClearScr = ESC + "2J"
	HideCur  = ESC + "?25l"
	ShowCur  = ESC + "?25h"
	Reset    = ESC + "0m"
)

// Colors
const (
	Green     = ESC + "32m"
	Yellow    = ESC + "33m"
	Cyan      = ESC + "36m"
	White     = ESC + "37m"
	BrRed     = ESC + "91m"
	BrGreen   = ESC + "92m"
	BrYellow  = ESC + "93m"
	BrMagenta = ESC + "95m"
	BrCyan    = ESC + "96m"

	WaterBg   = ESC + "48;5;18m"
	SandBg    = ESC + "48;5;94m"  // sand background
	SandFg    = ESC + "38;5;180m"
	SandFg2   = ESC + "38;5;186m" // lighter sand highlight
	DarkSand  = ESC + "38;5;137m"
	ShellFg   = ESC + "38;5;223m" // seashell color
	PebbleFg  = ESC + "38;5;245m" // gray pebbles
	StoneFg   = ESC + "38;5;249m" // rock color
	DarkGreen = ESC + "38;5;22m"
	LimeGreen = ESC + "38;5;118m"
	SeaGreen  = ESC + "38;5;35m"
	Orange    = ESC + "38;5;208m"
	Pink      = ESC + "38;5;213m"
	GoldFg    = ESC + "38;5;220m"
	CoralPink = ESC + "38;5;204m"
	CoralOrg  = ESC + "38;5;209m"
	CoralRed  = ESC + "38;5;196m"
	Plankton  = ESC + "38;5;60m"  // dim floating particle
)

// Direction a fish or crab faces
type Direction int

const (
	Left Direction = iota
	Right
)

// FishSpecies holds the ASCII art for one type of fish
type FishSpecies struct {
	Right []string
	Left  []string
	Color string
}

var fishSpecies = []FishSpecies{
	// Neon tetra
	{Right: []string{"><>"}, Left: []string{"<><"}, Color: BrCyan},
	// Clownfish
	{Right: []string{"><))°>"}, Left: []string{"<°((<>"}, Color: Orange},
	// Angelfish
	{
		Right: []string{`  /\  `, `>-)°> `, `  \/  `},
		Left:  []string{`  /\  `, ` <°(-<`, `  \/  `},
		Color: BrYellow,
	},
	// Pufferfish
	{
		Right: []string{` .--. `, `( °  >`, ` '--' `},
		Left:  []string{` .--. `, `<  ° )`, ` '--' `},
		Color: BrGreen,
	},
	// Betta
	{
		Right: []string{`  /)  `, `<{°)=>`, `  \)  `},
		Left:  []string{`  (\  `, `<=({°>`, `  (/  `},
		Color: BrMagenta,
	},
	// Swordtail
	{Right: []string{`--==><{{°>`}, Left: []string{`<°}}>==--`}, Color: BrRed},
	// Jellyfish
	{
		Right: []string{` .--. `, `( °° )`, ` )~~( `, ` |  | `},
		Left:  []string{` .--. `, `( °° )`, ` )~~( `, ` |  | `},
		Color: Pink,
	},
	// Goldfish
	{
		Right: []string{`  __  `, `>( °}>`, `  --  `},
		Left:  []string{`  __  `, `<{° )>`, `  --  `},
		Color: GoldFg,
	},
}

// Fish is a live fish swimming in the tank
type Fish struct {
	X, Y      int
	Species   *FishSpecies
	Dir       Direction
	Speed     int
	TickCount int
}

// Bubble floats upward
type Bubble struct{ X, Y int }

// Crab scuttles along the sand
type Crab struct {
	X, Y      int
	Dir       Direction
	Frame     int
	TickCount int
}

var crabRight = []string{"V(°°)V", "v(°°)v"}
var crabLeft  = []string{"V(°°)V", "v(°°)v"}

// Starfish sits on the sand
type Starfish struct{ X, Y int }

// Plant is seaweed that sways
type Plant struct {
	X     int
	Art   []string
	Color string
}

// Coral is a colorful reef formation
type Coral struct {
	X     int
	Art   []string
	Color string
}

// Rock sits on the sand
type Rock struct {
	X   int
	Art []string
}

// cell is one character in the render grid
type cell struct {
	ch    rune
	color string
}

var tankW, tankH int

// ---------------------------------------------------------------------------
// Terminal helpers
// ---------------------------------------------------------------------------

func getTerminalSize() (int, int) {
	type winsize struct{ Row, Col, Xpx, Ypx uint16 }
	ws := &winsize{}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno == 0 && ws.Col > 0 && ws.Row > 0 {
		return int(ws.Col), int(ws.Row)
	}
	return 80, 24
}

func inBounds(x, y int) bool {
	return x >= 0 && x < tankW && y >= 0 && y < tankH
}

// ---------------------------------------------------------------------------
// Spawning helpers
// ---------------------------------------------------------------------------

func spawnFish(n int) []Fish {
	out := make([]Fish, n)
	for i := range out {
		sp := &fishSpecies[rand.Intn(len(fishSpecies))]
		out[i] = Fish{
			X:       rand.Intn(tankW-20) + 5,
			Y:       rand.Intn(tankH-8) + 3,
			Species: sp,
			Dir:     Direction(rand.Intn(2)),
			Speed:   1 + rand.Intn(3),
		}
	}
	return out
}

func spawnCrabs() []Crab {
	n := 1 + rand.Intn(2)
	out := make([]Crab, n)
	for i := range out {
		out[i] = Crab{
			X:   rand.Intn(tankW-20) + 5,
			Y:   tankH - 3,
			Dir: Direction(rand.Intn(2)),
		}
	}
	return out
}

func generatePlants() []Plant {
	colors := []string{Green, BrGreen, DarkGreen, LimeGreen, SeaGreen}
	// Wide, lush seaweed shapes
	seaweedArts := [][]string{
		{" (} ", "  {)", " (} ", "  {)", "  || "},
		{"  )\\ ", " //(  ", "  )\\ ", " //(  ", "  ||  "},
		{" )(  ", "  ()  ", " )(  ", "  ()  ", "  )(  ", "  ||  "},
		{" })  ", " ({  ", " })  ", " ({  ", "  |   "},
		{"  })) ", " (({  ", "  })) ", " (({  ", "  })) ", "  ||  "},
		{" )\\ ", "//( ", " )\\ ", "//( ", " || "},
	}
	n := 6 + rand.Intn(5)
	out := make([]Plant, 0, n)
	for i := 0; i < n; i++ {
		base := seaweedArts[rand.Intn(len(seaweedArts))]
		// Vary height by trimming or repeating the top portion
		extra := rand.Intn(3)
		art := make([]string, 0, len(base)+extra)
		for e := 0; e < extra; e++ {
			art = append(art, base[e%2])
		}
		art = append(art, base...)
		out = append(out, Plant{X: 3 + rand.Intn(tankW-8), Art: art, Color: colors[rand.Intn(len(colors))]})
	}
	return out
}

func generateRocks() []Rock {
	shapes := [][]string{
		{` /~~\ `, `(    )`, ` \__/ `},
		{`  __  `, ` /  \ `, `/____\`},
		{` ___ `, `/   \`, `\___/`},
	}
	n := 2 + rand.Intn(3)
	out := make([]Rock, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Rock{X: 5 + rand.Intn(tankW-15), Art: shapes[rand.Intn(len(shapes))]})
	}
	return out
}

func generateStarfish() []Starfish {
	n := 1 + rand.Intn(3)
	out := make([]Starfish, n)
	for i := range out {
		out[i] = Starfish{X: 4 + rand.Intn(tankW-10), Y: tankH - 3}
	}
	return out
}

func generateCoral() []Coral {
	shapes := [][]string{
		{`  (@@)  `, ` (@@@@@)`, `  (@@)  `, `   ||   `},
		{` ,I, `, `(III)`, ` 'I' `},
		{`  oOo  `, ` oOOOo `, `  oOo  `},
		{` ^  ^ `, `(^^^^)`, ` ^^^^`},
	}
	colors := []string{CoralPink, CoralOrg, CoralRed, Pink}
	n := 2 + rand.Intn(3)
	out := make([]Coral, 0, n)
	for i := 0; i < n; i++ {
		shape := shapes[rand.Intn(len(shapes))]
		out = append(out, Coral{
			X:     6 + rand.Intn(tankW-16),
			Art:   shape,
			Color: colors[rand.Intn(len(colors))],
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// Movement
// ---------------------------------------------------------------------------

func moveFish(f *Fish) {
	if f.Dir == Right {
		f.X++
		if f.X > tankW-2 {
			f.Dir = Left
			f.Y = clamp(3+rand.Intn(tankH-9), 3, tankH-6)
		}
	} else {
		f.X--
		if f.X < -len(f.Species.Left[0]) {
			f.Dir = Right
			f.Y = clamp(3+rand.Intn(tankH-9), 3, tankH-6)
		}
	}
	// Slight vertical drift
	if rand.Intn(12) == 0 {
		f.Y = clamp(f.Y+rand.Intn(3)-1, 3, tankH-6-len(f.Species.Right))
	}
}

func clamp(v, lo, hi int) int {
	if lo > hi {
		lo = hi
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

func drawSprite(grid [][]cell, x, y int, art []string, color string, skipSpace bool) {
	for dy, line := range art {
		for dx, ch := range line {
			px, py := x+dx, y+dy
			if inBounds(px, py) && (!skipSpace || ch != ' ') {
				grid[py][px] = cell{ch: ch, color: color}
			}
		}
	}
}

func render(buf *strings.Builder, grid [][]cell, fishes []Fish, bubbles []Bubble, plants []Plant, rocks []Rock, corals []Coral, starfish []Starfish, crabs []Crab, chestX, tick int) {
	buf.Reset()
	buf.Grow(tankW * tankH * 6)

	buf.WriteString(ESC + "H") // cursor home

	// Row 1: animated wave surface
	buf.WriteString(WaterBg + BrCyan)
	for x := 0; x < tankW; x++ {
		switch (x + tick/2) % 12 {
		case 0, 1, 2:
			buf.WriteRune('~')
		case 3, 4, 5:
			buf.WriteString("≈")
		case 6, 7, 8:
			buf.WriteRune('~')
		default:
			buf.WriteString("∽")
		}
	}

	// Row 2: sub-surface ripple
	buf.WriteString(Reset + WaterBg + Cyan)
	for x := 0; x < tankW; x++ {
		if (x+tick/3+4)%8 < 4 {
			buf.WriteRune('~')
		} else {
			buf.WriteRune(' ')
		}
	}

	// Clear grid
	for y := range grid {
		for x := range grid[y] {
			grid[y][x] = cell{ch: ' '}
		}
	}

	// Sand (bottom 2 rows) with shells, pebbles, and varied texture
	sandRow := tankH - 2
	for x := 0; x < tankW; x++ {
		hash := (x*7 + 13) % 31 // deterministic variety
		switch {
		case hash == 0:
			grid[sandRow][x] = cell{ch: '@', color: ShellFg} // shell
		case hash == 5:
			grid[sandRow][x] = cell{ch: 'o', color: PebbleFg} // pebble
		case hash == 12:
			grid[sandRow][x] = cell{ch: '~', color: ShellFg} // small shell
		case (x+tick/4)%5 == 0:
			grid[sandRow][x] = cell{ch: '.', color: DarkSand}
		case (x+3)%7 == 0:
			grid[sandRow][x] = cell{ch: ',', color: SandFg2}
		default:
			grid[sandRow][x] = cell{ch: '▒', color: SandFg}
		}
		if sandRow+1 < tankH {
			grid[sandRow+1][x] = cell{ch: '▓', color: SandFg}
		}
	}

	// Rocks
	for _, r := range rocks {
	drawSprite(grid, r.X, tankH-3-len(r.Art)+1, r.Art, StoneFg, false)
	}

	// Treasure chest
	chestArt := []string{` ___.____ `, `|  o$$o  |`, `|  $$$$  |`, `|________|`}
	for dy, line := range chestArt {
		y := tankH - 5 + dy
		for dx, ch := range line {
			x := chestX + dx
			if inBounds(x, y) {
				color := Yellow
				if ch == '$' {
					color = GoldFg
				}
				grid[y][x] = cell{ch: ch, color: color}
			}
		}
	}

	// Plants with sway animation
	for _, p := range plants {
		for dy, line := range p.Art {
			y := tankH - 3 - len(p.Art) + dy
			sway := 0
			if (tick/6+dy)%4 < 2 {
				sway = 1
			}
			for dx, ch := range line {
				x := p.X + dx + sway
				if inBounds(x, y) {
					grid[y][x] = cell{ch: ch, color: p.Color}
				}
			}
		}
	}

	// Coral formations
	for _, c := range corals {
		drawSprite(grid, c.X, tankH-3-len(c.Art)+1, c.Art, c.Color, true)
	}

	// Starfish
	sfArt := []string{`\|/`, `-*-`, `/|\`}
	for _, sf := range starfish {
		drawSprite(grid, sf.X, sf.Y-2, sfArt, Orange, false)
	}

	// Crabs
	for _, c := range crabs {
		frames := crabRight
		if c.Dir == Left {
			frames = crabLeft
		}
		for dx, ch := range frames[c.Frame] {
			x := c.X + dx
			if inBounds(x, c.Y) {
				grid[c.Y][x] = cell{ch: ch, color: BrRed}
			}
		}
	}

	// Fish
	for _, f := range fishes {
		art := f.Species.Right
		if f.Dir == Left {
			art = f.Species.Left
		}
		drawSprite(grid, f.X, f.Y, art, f.Species.Color, true)
	}

	// Ambient particles / plankton
	for y := 3; y < tankH-3; y++ {
		for x := 0; x < tankW; x++ {
			if grid[y][x].ch == ' ' {
				hash := (x*131 + y*97 + tick/8) % 197
				if hash == 0 {
					grid[y][x] = cell{ch: '.', color: Plankton}
				} else if hash == 42 {
					grid[y][x] = cell{ch: '·', color: Plankton}
				}
			}
		}
	}


	// Bubbles (topmost layer)
	for _, b := range bubbles {
		if inBounds(b.X, b.Y) {
			ch := '°'
			if b.Y < 5 {
				ch = 'o'
			}
			if b.Y < 3 {
				ch = 'O'
			}
			grid[b.Y][b.X] = cell{ch: ch, color: BrCyan}
		}
	}

	// Emit grid (rows 2..tankH-1)
	for y := 2; y < tankH; y++ {
		if y >= tankH-2 {
			buf.WriteString(Reset + SandBg)
		} else {
			buf.WriteString(Reset + WaterBg)
		}
		lastColor := ""
		for x := 0; x < tankW; x++ {
			c := grid[y][x]
			if c.color != lastColor {
				buf.WriteString(c.color)
				lastColor = c.color
			}
			buf.WriteRune(c.ch)
		}
	}
	buf.WriteString(Reset)
	fmt.Fprint(os.Stdout, buf.String())
}

// ---------------------------------------------------------------------------
// Main loop
// ---------------------------------------------------------------------------

func main() {
	rand.Seed(time.Now().UnixNano())

	tankW, tankH = getTerminalSize()
	if tankW < 40 || tankH < 15 {
		fmt.Println("Terminal too small! Need at least 40x15.")
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)

	fmt.Print(HideCur + ClearScr)
	defer fmt.Print(ShowCur + Reset + ClearScr)

	fishes := spawnFish(12 + rand.Intn(6))
	bubbles := make([]Bubble, 0, 30)
	plants := generatePlants()
	rocks := generateRocks()
	corals := generateCoral()
	stars := generateStarfish()
	crabs := spawnCrabs()
	chestX := tankW/4 + rand.Intn(tankW/3)

	var buf strings.Builder
	grid := make([][]cell, tankH)
	for i := range grid {
		grid[i] = make([]cell, tankW)
	}

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	tick := 0
	for {
		select {
		case sig := <-sigChan:
			if sig == syscall.SIGWINCH {
				tankW, tankH = getTerminalSize()
				grid = make([][]cell, tankH)
				for i := range grid {
					grid[i] = make([]cell, tankW)
				}
				plants = generatePlants()
				rocks = generateRocks()
				fmt.Print(ClearScr)
				continue
			}
			return

		case <-ticker.C:
			tick++

			// Move fish
			for i := range fishes {
				fishes[i].TickCount++
				if fishes[i].TickCount >= fishes[i].Speed {
					fishes[i].TickCount = 0
					moveFish(&fishes[i])
				}
			}

			// Move crabs
			for i := range crabs {
				crabs[i].TickCount++
				if crabs[i].TickCount >= 4 {
					crabs[i].TickCount = 0
					crabs[i].Frame = (crabs[i].Frame + 1) % 2
					if rand.Intn(10) < 3 {
						if crabs[i].Dir == Right {
							crabs[i].X++
						} else {
							crabs[i].X--
						}
						if crabs[i].X <= 2 || crabs[i].X >= tankW-12 {
							crabs[i].Dir = 1 - crabs[i].Dir
						}
					}
				}
			}

			// Bubble spawning from fish
			if tick%5 == 0 {
				for _, f := range fishes {
					if rand.Intn(8) == 0 {
						bx := f.X
						if f.Dir == Right {
							bx += len(f.Species.Right[0])
						}
						bubbles = append(bubbles, Bubble{X: bx, Y: f.Y})
					}
				}
			}

			// Chest bubbles
			if tick%7 == 0 && rand.Intn(3) == 0 {
				bubbles = append(bubbles, Bubble{X: chestX + 3, Y: tankH - 5})
			}

			// Rise bubbles
			if tick%3 == 0 {
				alive := bubbles[:0]
				for _, b := range bubbles {
					b.Y--
					if rand.Intn(3) == 0 {
						b.X += rand.Intn(3) - 1
					}
					if b.Y > 2 {
						alive = append(alive, b)
					}
				}
				bubbles = alive
			}

			render(&buf, grid, fishes, bubbles, plants, rocks, corals, stars, crabs, chestX, tick)
		}
	}
}
