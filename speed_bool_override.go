// An implementation of Conway's Game of Life.
// See : http://golang.org/doc/play/life.go

// go run speed_bool.go 


package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

/*
type Fielder interface {
 //NewField(w, h int) *Field
 UniformRandom(pct float32)
 LoadString(s string)
 //Set(x, y int, b bool)
 //isSet(x, y int) bool
 //IterateCell(x, y int) bool
 String() string 
}
*/

type FieldHighLeveler interface {
 UniformRandom(pct float32)
 LoadString(s string)
 String() string 
 
 Iterate(f FieldHighLeveler) 
 IterateCell(x, y int) bool
 
 Set(x, y int, b bool)
 isSet(x, y int) bool
 
 get_height() int
 get_width() int
}

type FieldHighLevel struct {
}

// Field represents a two-dimensional field of cells.
type Field_BoolArray struct {
	s    [][]bool
	w, h int
	
	FieldHighLevel
}

// NewField_BoolArray returns an empty field of the specified width and height.
func NewField_BoolArray(w, h int) *Field_BoolArray {
	s := make([][]bool, h)
	for i := range s {
		s[i] = make([]bool, w)
	}
	return &Field_BoolArray{s: s, w: w, h: h}
}

// puts Field in a random state
func (f *FieldHighLevel) UniformRandom(pct float32) {
	for y := 0; y < f.get_height(); y++ {
		for x := 0; x < f.get_width(); x++ {
			f.Set(x, y, (rand.Float32()<pct))
		}
	}
}

// loads Field from a string : Using '\n' and 'X' as markers
func (f *FieldHighLevel) LoadString(s string) {
 x:=0
 y:=0
 for _, v := range s[:] {
  if v == 'X' {
   f.Set(x,y, true)
  } 
  x+=1
  if v == '\n' {
   x=0
   y+=1
  } 
 }
}

func (f *FieldHighLevel) Set(x, y int, b bool) {
	fmt.Print("FieldHighLevel.Set stub called\n") 
}
func (f *FieldHighLevel) isSet(x, y int) bool {
	fmt.Print("FieldHighLevel.isSet stub called\n") 
	return false;
}
func (f *FieldHighLevel) get_height() int {
	fmt.Print("FieldHighLevel.get_height stub called\n") 
	return 0;
}
func (f *FieldHighLevel) get_width() int {
	fmt.Print("FieldHighLevel.get_width stub called\n") 
	return 0;
}

// Set sets the state of the specified cell to the given value.
func (f *Field_BoolArray) Set(x, y int, b bool) {
	f.s[y][x] = b
}

func (f *Field_BoolArray) get_height() int {
	return f.h;
}

func (f *Field_BoolArray) get_width() int {
	return f.w;
}


// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
func (f *Field_BoolArray) isSet_toroid(x, y int) bool {
	x += f.w
	x %= f.w
	y += f.h
	y %= f.h
	return f.s[y][x]
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are not wrapped
func (f *Field_BoolArray) isSet(x, y int) bool {
 if(x<0 || x>=f.w) {
  return false
 }
 if(y<0 || y>=f.h) {
  return false
 }
	return f.s[y][x]
}

// Next returns the state of the specified cell at the next time step.
func (f *FieldHighLevel) IterateCell(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && f.isSet(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && f.isSet(x, y)
}

func (f *FieldHighLevel) Iterate(next FieldHighLeveler) {
	// Update the state of the next field (next) in-place from the current field (f).
	for y := 0; y < f.get_height(); y++ {
		for x := 0; x < f.get_width(); x++ {
			next.Set(x, y, f.IterateCell(x, y))
		}
	}
}

// String returns the game board as a string.
func (f *FieldHighLevel) String() string {
	var buf bytes.Buffer
 outer:=1;
	for y := 0-outer; y < f.get_height()+outer; y++ {
		for x := 0-outer; x < f.get_width()+outer; x++ {
			b := byte('-')
			if f.isSet(x, y) {
				b = '*'
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}


// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b FieldHighLeveler
}

// NewLife returns a new Life game state 
func NewLife(w, h int) *Life {
	return &Life{
		a: NewField_BoolArray(w, h), b: NewField_BoolArray(w, h),
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).

    l.a.Iterate(l.b)
 
	// Now swap fields a and b.
	l.a, l.b = l.b, l.a
}

func main_orig() {
	l := NewLife(20, 20)
	for i := 0; i < 65; i++ {
		l.Step()
		fmt.Print("\x0c", l) // Clear screen and print field.
		time.Sleep(time.Second / 30)
	}
}

func main_() {
	l := NewLife(20, 20)
//	l.UniformRandom(.2)
 
	const glider = `
--X
X-X
-XX`

 l.a.LoadString(glider[1:])
 
	//fmt.Print("\x0c", l) 
	fmt.Print(l.a, "\n") 
	for i := 0; i < 65; i++ {
		l.Step()
	}
	fmt.Print(l.a, "\n") 
 
/* 
	for i := 0; i < 65; i++ {
		l.Step()
		fmt.Print("\x0c", l) // Clear screen and print field.
		time.Sleep(time.Second / 30)
	}
*/
}

func main() {
	const glider = `
--X
X-X
-XX`

 start := time.Now()

  l := NewLife(20, 20)
  l.a.LoadString(glider[1:])
  
  return;

 for iter:=0; iter<1000; iter++ {
  l := NewLife(20, 20)
  l.a.LoadString(glider[1:])
  
  for i := 0; i < 65; i++ {
   l.Step()
  }
 } 
 
 elapsed := time.Since(start)
 fmt.Printf("1000 iterations took %s\n", elapsed)
}
