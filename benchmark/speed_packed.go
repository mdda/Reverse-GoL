// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"fmt"
 "time"
)

const board_width  int =20
const board_height int =20

// Board represents a two-dimensional field of cells.
type Board_BoolPacked struct {
	s    []int32
	h,w  int // Only used for GENERIC functions
}

var count_bits_array [512]byte

func build_count_bits_array() {
	for x := 0; x < 512; x++ {
		cnt := byte(0)
		for i := uint(0); i < 9; i++ {
			if x&(1<<i) > 0 {
				cnt++
			}
		}
		count_bits_array[x] = cnt
	}
	//fmt.Print(count_bits_array, "\n")
}

var board_empty *Board_BoolPacked

// NewBoard_BoolArray returns an empty field of the specified width and height.
func NewBoard_BoolPacked(w,h int) *Board_BoolPacked { // OPTIMIZED FOR BoolPacked
	if board_width > 22 {
		fmt.Print("TOO LARGE AN ARRAY for bottom 3 bytes of int32!\n")
	}
	
	s := make([]int32, board_height+2) // Need padding before and after
	return &Board_BoolPacked{s: s, h:board_height, w:board_width}
}

func (dest *Board_BoolPacked) CopyFrom(src *Board_BoolPacked) { // OPTIMIZED FOR BoolPacked
	dest.s = make([]int32, board_height+2)
	for y := 0; y<board_height+2; y++ {
		dest.s[y] = src.s[y]
	}
}

// Set sets the state of the specified cell to the given value.
func (f *Board_BoolPacked) Set(x, y int, b bool) { // OPTIMIZED FOR BoolPacked
	//  The (+1,+1) offsets are to account for the zeroed-out borders
	if b { //  This is a set=TRUE
		f.s[y+1] |= (1 << uint(x+1))
	} else { //  This is a set=FALSE
		f.s[y+1] &= ^(1 << uint(x+1))
	}
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are not wrapped
func (f *Board_BoolPacked) isSet(x, y int) bool { // OPTIMIZED FOR BoolPacked
	return (f.s[y+1] & (1 << uint(x+1))) != 0
}

// Update the state of the next field (next) in-place from the current field (f).
func (f *Board_BoolPacked) Iterate(next *Board_BoolPacked) { // OPTIMIZED FOR BoolPacked
	// This is done rather over-efficiently...

	// These are constants - the game bits pass over them
	top_filter := int32(7) //  111
	mid_filter := int32(5) //  101
	bot_filter := int32(7) //  111

	current_filter := int32(2) //  010

	next.s[0] = 0
	for r := 1; r <= board_height; r++ {
		r_top := f.s[r-1]
		r_mid := f.s[r]
		r_bot := f.s[r+1]

		acc := int32(0)
		p := int32(2) // Start in the middle row, one column in (000000010b)

		for c := 1; c <= board_width; c++ {
			cnt := count_bits_array[((r_top&top_filter)<<6)|
									((r_mid&mid_filter)<<3)|
									((r_bot&bot_filter))    ]

			// if 1==1 { acc |= p }  // Check bit-twiddling bounds

			// Return next state according to the game rules:
			//  exactly 3 neighbors: on,
			//  exactly 2 neighbors: maintain current state,
			//  otherwise: off.
			//  return alive == 3 || alive == 2 && f.Alive(x, y)

			if (cnt == 3) || (cnt == 2 && ((r_mid&current_filter) != 0)) {
				acc |= p
			}

			// Move the 'setting-bit' over
			p <<= 1

			// Shift the arrays over into base filterable position
			r_top >>= 1
			r_mid >>= 1
			r_bot >>= 1
		}
		next.s[r] = acc
	}
	next.s[board_height+1] = 0
}

// loads Field from a string : Using '\n' and 'X' as markers
func (f *Board_BoolPacked) LoadString(s string) {
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


func init() {
	fmt.Print("init() called\n")
	build_count_bits_array()
	board_empty = NewBoard_BoolPacked(board_width, board_height)
}

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b *Board_BoolPacked
}

// NewLife returns a new Life game state 
func NewLife(w, h int) *Life {
	return &Life{
		a: NewBoard_BoolPacked(w,h), b: NewBoard_BoolPacked(w,h),
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).

    l.a.Iterate(l.b)
 
	// Now swap fields a and b.
	l.a, l.b = l.b, l.a
}


func main() {
	const glider = `
--X
X-X
-XX`

 start := time.Now()

//  l := NewLife(20, 20)
//  l.a.LoadString(glider[1:])
  
//  return;

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
