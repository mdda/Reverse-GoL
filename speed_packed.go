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

// Field represents a two-dimensional field of cells.
type Field_BoolPacked struct {
	s    []int32
	w, h int
}


var count_bits_array [512]byte

func build_count_bits_array() {
	for x:=0; x<512; x++ {
		cnt:=byte(0)
		for i:=uint(0); i<9; i++ {
			if x & (1<<i) > 0 {
				cnt++
			}
		}
		count_bits_array[x]=cnt;
	}
	//fmt.Print(count_bits_array, "\n") 
}

// NewField_BoolArray returns an empty field of the specified width and height.
func NewField_BoolPacked(w, h int) *Field_BoolPacked {
	if(w>29) {
      fmt.Print("TOO LARGE AN ARRAY!\n") 
	}
	s := make([]int32, h+2)  // Need padding before and after
	return &Field_BoolPacked{s: s, w: w, h: h}
}

// puts Field in a random state
func (f *Field_BoolPacked) UniformRandom(pct float32) {
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			f.Set(x, y, (rand.Float32()<pct))
		}
	}
}

// loads Field from a string : Using '\n' and 'X' as markers
func (f *Field_BoolPacked) LoadString(s string) {
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

// Set sets the state of the specified cell to the given value.
func (f *Field_BoolPacked) Set(x, y int, b bool) {
//	f.s[y][x] = b
//  The (+1,+1) offsets are to account for the zeroed-out borders
  if(b) {  //  This is a set=TRUE
    f.s[y+1] |= (1<<uint(x+1))
  } else {   //  This is a set=FALSE
    f.s[y+1] &= ^(1<<uint(x+1))
  }
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are not wrapped
func (f *Field_BoolPacked) isSet(x, y int) bool {
 return (f.s[y+1] & (1<<uint(x+1))) != 0
}

// Next returns the state of the specified cell at the next time step.
func (f *Field_BoolPacked) IterateCell(x, y int) bool {
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

func (f *Field_BoolPacked) Iterate_Generic(next *Field_BoolPacked) {
	// Update the state of the next field (next) in-place from the current field (f).
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			next.Set(x, y, f.IterateCell(x, y))
		}
	}
}

func (f *Field_BoolPacked) Iterate(next *Field_BoolPacked) {
	// Update the state of the next field (next) in-place from the current field (f).
	
	// This is done rather over-efficiently...
	
    // These are constants - the game bits pass over them
    top_filter     := int32(7)  //  111
    mid_filter     := int32(5)  //  101
    bot_filter     := int32(7)  //  111
    
    current_filter := int32(2)  //  010
            
    next.s[0]=0
    for r:=1; r<=f.h; r++ {
		r_top := f.s[r-1]
		r_mid := f.s[r]
		r_bot := f.s[r+1]
            
		acc := int32(0)
		p   := int32(2) // Start in the middle row, one column in (000000010b)
            
        for c:=1; c<=f.w; c++ {
            cnt := count_bits_array[
                  ((r_top & top_filter) << 6 ) |
                  ((r_mid & mid_filter) << 3 ) |
                   (r_bot & bot_filter) ]
            
			// if 1==1 { acc |= p }  // Check bit-twiddling bounds
			
			// Return next state according to the game rules:
			//  exactly 3 neighbors: on,
			//  exactly 2 neighbors: maintain current state,
			//  otherwise: off.
			//  return alive == 3 || alive == 2 && f.Alive(x, y)
			
			if (cnt==3) || (cnt==2 && (r_mid & current_filter)!=0) {
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
    next.s[f.h+1]=0
}

// String returns the game board as a string.
func (f *Field_BoolPacked) String() string {
	var buf bytes.Buffer
 outer:=1;
	for y := 0-outer; y < f.h+outer; y++ {
		for x := 0-outer; x < f.w+outer; x++ {
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
	a, b *Field_BoolPacked
}

// NewLife returns a new Life game state 
func NewLife(w, h int) *Life {
	if count_bits_array[1] != 1 {
		build_count_bits_array()
	}
	return &Life{
		a: NewField_BoolPacked(w, h), b: NewField_BoolPacked(w, h),
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

func main() {
	const glider = `
--X
X-X
-XX`

	for pct := float32(0.1); pct<1.0; pct += 0.1 {
		l := NewLife(20, 20)
		//l.a.LoadString(glider[1:])
	    l.a.UniformRandom(pct + 0.01)
 
		//fmt.Print("\x0c", l) 
		fmt.Print(l.a, "\n") 
		for i := 0; i < 5; i++ {   // 65 max reasonable steps for glider
			l.Step()
		}
		fmt.Print(l.a, "\n") 
	}
}

func main_() {
	const glider = `
--X
X-X
-XX`

 start := time.Now()

 for iter:=0; iter<1000; iter++ {
  l := NewLife(20, 20)
  l.a.LoadString(glider[1:])
  
  for i := 0; i < 65; i++ {
   l.Step()
  }
  if(iter==0) {
   fmt.Print(l.a, "\n") 
  }
 } 
 
 elapsed := time.Since(start)
 fmt.Printf("1000 iterations took %s\n", elapsed)
}
