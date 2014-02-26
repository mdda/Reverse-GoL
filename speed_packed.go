// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"fmt"
	"math/rand"
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

func (attempt *Board_BoolPacked) CompareTo(target *Board_BoolPacked, diff *Board_BoolPacked) int { // OPTIMIZED FOR BoolPacked
	r := 0
	match := int32(0)
	lowest_byte := int32(0xff)
	for y := 1; y<=board_height; y++ {
		match = attempt.s[y] ^ target.s[y] // This covers all of lower 24 bits, which is what we care about
		if match>0 {
			r += int(count_bits_array[(match>>0) & lowest_byte] + 
					 count_bits_array[(match>>8) & lowest_byte] + 
					 count_bits_array[(match>>16) & lowest_byte])
		}
		if diff != nil {
			diff.s[y]=match
		}
	}
	return r
}

type Patch int

func (f *Board_BoolPacked) MakePatch(x,y int) Patch {
	var p Patch=0
	for dy:=-2; dy<=+2; dy++ {
		for dx:=-2; dx<=+2; dx++ {
			p<<=1
			if f.isSet_safe(x+dx,y+dy) {
				p |= 1
			}
		}
	}
	return p
}




func (f *Board_BoolPacked) MutateFlipBits(count int) {
	for c:=0; c<count; c++ {
		// Pick two random locations, and copy the bit from one to the other
		src_x, src_y := rand.Intn(f.w), rand.Intn(f.h)
		dst_x, dst_y := rand.Intn(f.w), rand.Intn(f.h)
		
		f.Set(dst_x, dst_y, f.isSet(src_x, src_y))
	}
}

func CoordWithinRadius(origin int, dim int, radius int) int {
	q :=-1
	for ; (q<0 || q>=dim); q = origin+rand.Intn(radius*2+1)-radius {
	}
	return q
}

func (f *Board_BoolPacked) MutateRadiusBits_SwitchARoo(another_mutation_pct, radius int) {
	// Pick a random location
	src_x, src_y := rand.Intn(f.w), rand.Intn(f.h)
	for {
		if rand.Intn(100)>another_mutation_pct {
			break
		}
			
		// and another within L1(radius) of it
		dst_x := CoordWithinRadius(src_x, f.w, radius)
		dst_y := CoordWithinRadius(src_y, f.h, radius)

		src_isSet := f.isSet(src_x, src_y)
		// Switch-a-roo
		f.Set(src_x, src_y, f.isSet(dst_x, dst_y))
		f.Set(dst_x, dst_y, src_isSet)
	}
}

func (f *Board_BoolPacked) MutateRadiusBits(another_mutation_pct, radius int) {
	// Pick a random location
	src_x, src_y := rand.Intn(f.w), rand.Intn(f.h)
	for {
		// Pick an L1 radius
		r_up := rand.Intn(radius)
		r_down := rand.Intn(radius)
		for x:=src_x-r_down; x<=src_x+r_up; x++ {
			for y:=src_y-r_down; y<=src_y+r_up; y++ {
				if 0<=x && x<f.w && 0<=y && y<f.h {
					f.Set(x, y, f.isSet(x, y) != true)
				}
			}
		}
		if rand.Intn(100)>another_mutation_pct {
			break
		}
	}
}

func (f *Board_BoolPacked) MutateMask(mask *Board_BoolPacked, another_mutation_pct, radius int) { // OPTIMIZED FOR BoolPacked
	// This isn't really a uniform picker amongst mask bits, but it makes an effort to be fast...
	for {
		// Pick a random row, and find the first line there (or after) that has a non-zero in it
		y := rand.Intn(board_height)
		for cnt := board_height; (mask.s[y+1]==0) && cnt>0; cnt-- {
			//fmt.Printf("MutateMask moving to next line %2d (count=%2d)\n", y, cnt)
			y++
			if y>=board_height {
				//fmt.Printf("MutateMask wraparound after line %2d\n", y)
				y=0
			}
		}
		mask_row := mask.s[y+1]
		if mask_row==0 {
			// We looped around : No mask>0 => No mask to be found.  i.e. no mutation to do
			//fmt.Printf("MutateMask no diffs : Perfect! on line %d\n")
			//fmt.Println(mask)
			break
		}
		
		// Pick a random column
		x := rand.Intn(board_width)
		for cnt := board_width; ((mask_row & (1<<uint(x+1)))==0) && cnt>0; cnt-- {
			x++
			if x>=board_width {
				x=0
			}
		}
		
		// Have found an x,y
		if mask.isSet(x,y) != true {
			fmt.Printf("MutateMask bit-twiddle failure %22b @ %2d=%22b %d\n", mask_row, x+1)
			return
		}
		
		//f.Set(x,y, f.isSet(x,y)==false) // Flip the bit which corresponds to the diff
		
		x_offset := CoordWithinRadius(x, board_width, radius)
		y_offset := CoordWithinRadius(y, board_height, radius)
		f.Set(x_offset,y_offset, f.isSet(x_offset,y_offset)==false) // Flip the bit which corresponds to the diff+/-a radius distance
		
		//fmt.Printf("MutateMask flip bit (%2d,%2d)\n", x,y)
		if rand.Intn(100)>another_mutation_pct {
			break
		}
		//fmt.Printf("MutateMask round again\n")
	}
}



func (offspring *Board_BoolPacked) CrossoverFrom_Horizontal(p1, p2 *Board_BoolPacked) { // OPTIMIZED FOR BoolPacked
	offspring.s = make([]int32, board_height+2)
	cross := rand.Intn(board_height+2)
	for y := 0; y<board_height+2; y++ {
		if false && y<cross {
			offspring.s[y] = p1.s[y]
		} else {
			offspring.s[y] = p2.s[y]
		}
	}
/*
	// offspring.CopyFrom(p1)
	if(rand.Intn(100)>50) {
		// Horizontal dividing line
		
	} else {
		// Vertical dividing line
	}
*/
}

func (offspring *Board_BoolPacked) CrossoverFrom(p1, p2 *Board_BoolPacked) {
	offspring.CopyFrom(p1) // Grab p1 ASAP
	
	// Pick a random location
	src_x, src_y := rand.Intn(offspring.w), rand.Intn(offspring.h)
	
	radius := 5
	// Pick an L1 radius
	r_up := rand.Intn(radius)
	r_down := rand.Intn(radius)
	
	// Copy the rectangular blog from p2
	for x:=src_x-r_down; x<=src_x+r_up; x++ {
		for y:=src_y-r_down; y<=src_y+r_up; y++ {
			if 0<=x && x<offspring.w && 0<=y && y<offspring.h {
				offspring.Set(x, y, p2.isSet(x, y))
			}
		}
	}
}


func init() {
	fmt.Print("init() called\n")
	build_count_bits_array()
}

