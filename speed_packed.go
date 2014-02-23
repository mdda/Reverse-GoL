// An implementation of Conway's Game of Life.

// go run speed_packed.go

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// Board represents a two-dimensional field of cells.
type Board_BoolPacked struct {
	s    []int32
	w, h int
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
func NewBoard_BoolPacked(w, h int) *Board_BoolPacked {
	if w > 29 {
		fmt.Print("TOO LARGE AN ARRAY for int32!\n")
	}
	s := make([]int32, h+2) // Need padding before and after
	return &Board_BoolPacked{s: s, w: w, h: h}
}

// puts Board in a random state
func (f *Board_BoolPacked) UniformRandom(pct float32) {
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			f.Set(x, y, (rand.Float32() < pct))
		}
	}
}

// loads Board from a string : Using '\n' and 'X' as markers
func (f *Board_BoolPacked) LoadString(s string) {
	x := 0
	y := 0
	for _, v := range s[:] {
		if v == 'X' {
			f.Set(x, y, true)
		}
		x++
		if v == '\n' {
			x = 0
			y++
		}
	}
}

func (f *Board_BoolPacked) LoadArray(csv_strings []string) {
	x := 0
	y := 0

	for _, v := range csv_strings[:] {
		if v == "1" {
			f.Set(x, y, true)
			//fmt.Print("*")
		}
		x++
		if x >= f.w {
			x = 0
			y++
		}
	}
}

// Set sets the state of the specified cell to the given value.
func (f *Board_BoolPacked) Set(x, y int, b bool) {
	//	f.s[y][x] = b
	//  The (+1,+1) offsets are to account for the zeroed-out borders
	if b { //  This is a set=TRUE
		f.s[y+1] |= (1 << uint(x+1))
	} else { //  This is a set=FALSE
		f.s[y+1] &= ^(1 << uint(x+1))
	}
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are not wrapped
func (f *Board_BoolPacked) isSet(x, y int) bool {
	return (f.s[y+1] & (1 << uint(x+1))) != 0
}

// Next returns the state of the specified cell at the next time step.
func (f *Board_BoolPacked) IterateCell(x, y int) bool {
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

func (f *Board_BoolPacked) Iterate_Generic(next *Board_BoolPacked) {
	// Update the state of the next field (next) in-place from the current field (f).
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			next.Set(x, y, f.IterateCell(x, y))
		}
	}
}

func (f *Board_BoolPacked) Iterate(next *Board_BoolPacked) {
	// Update the state of the next field (next) in-place from the current field (f).

	// This is done rather over-efficiently...

	// These are constants - the game bits pass over them
	top_filter := int32(7) //  111
	mid_filter := int32(5) //  101
	bot_filter := int32(7) //  111

	current_filter := int32(2) //  010

	next.s[0] = 0
	for r := 1; r <= f.h; r++ {
		r_top := f.s[r-1]
		r_mid := f.s[r]
		r_bot := f.s[r+1]

		acc := int32(0)
		p := int32(2) // Start in the middle row, one column in (000000010b)

		for c := 1; c <= f.w; c++ {
			cnt := count_bits_array[((r_top&top_filter)<<6)|
				((r_mid&mid_filter)<<3)|
				(r_bot&bot_filter)]

			// if 1==1 { acc |= p }  // Check bit-twiddling bounds

			// Return next state according to the game rules:
			//  exactly 3 neighbors: on,
			//  exactly 2 neighbors: maintain current state,
			//  otherwise: off.
			//  return alive == 3 || alive == 2 && f.Alive(x, y)

			if (cnt == 3) || (cnt == 2 && (r_mid&current_filter) != 0) {
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
	next.s[f.h+1] = 0
}

// String returns the game board as a string.
func (f *Board_BoolPacked) String() string {
	var buf bytes.Buffer
	outer := 1
	for y := 0 - outer; y < f.h+outer; y++ {
		for x := 0 - outer; x < f.w+outer; x++ {
			b := byte('-')
			if x < 0 || x >= f.w || y < 0 || y >= f.h {
				b = '0'
			}
			if f.isSet(x, y) {
				b = '*'
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (f *Board_BoolPacked) AddToStats(bs *BoardStats) {
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			if f.isSet(x, y) {
				bs.freq[x][y]++
			}
		}
	}
	bs.count++
}

type BoardStats struct {
	freq  [][]int
	w, h  int
	count int
}

// NewField_BoolArray returns an empty field of the specified width and height.
func NewBoardStats(w, h int) *BoardStats {
	freq := make([][]int, h)
	for i := range freq {
		freq[i] = make([]int, w)
	}
	//fmt.Print("CreatedBoardStats\n")
	return &BoardStats{freq: freq, w: w, h: h, count: 0}
}

// BoardIterator stores the state of a round of Conway's Game of Life.
type BoardIterator struct {
	current, temp_internal_only *Board_BoolPacked
}

// BoardIterator returns a new Life game state
func NewBoardIterator(w, h int) *BoardIterator {
	if count_bits_array[1] != 1 {
		build_count_bits_array()
	}
	return &BoardIterator{
		current: NewBoard_BoolPacked(w, h), temp_internal_only: NewBoard_BoolPacked(w, h),
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (bi *BoardIterator) Iterate(n int) {
	for i := 0; i < n; i++ {
		bi.current.Iterate(bi.temp_internal_only)
		// Now swap boards, to put the result in prime position
		bi.current, bi.temp_internal_only = bi.temp_internal_only, bi.current
	}
}

func main_orig() {
	l := NewBoardIterator(20, 20)
	for i := 0; i < 65; i++ {
		l.Iterate(1)
		fmt.Print("\x0c", l) // Clear screen and print field.
		time.Sleep(time.Second / 30)
	}
}

func main_test_random() {
	for pct := float32(0.1); pct < 1.0; pct += 0.1 {
		l := NewBoardIterator(20, 20)
		l.current.UniformRandom(pct + 0.01)

		//fmt.Print("\x0c", l)
		fmt.Print(l.current, "\n")
		l.Iterate(5) // 65 max reasonable steps for glider
		fmt.Print(l.current, "\n")
	}
}

func main_timer() {
	const glider = `
--X
X-X
-XX`

	start := time.Now()

	for iter := 0; iter < 1000; iter++ {
		l := NewBoardIterator(20, 20)
		l.current.LoadString(glider[1:])

		l.Iterate(65)

		if iter == 0 {
			fmt.Print(l.current, "\n")
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("1000 iterations took %s\n", elapsed)
}

func main_loader() {
	for pct := float32(0.1); pct < 1.0; pct += 0.1 {
		l := NewBoardIterator(20, 20)
		l.current.UniformRandom(pct + 0.01)

		//fmt.Print("\x0c", l)
		fmt.Print(l.current, "\n")
		l.Iterate(5)
		fmt.Print(l.current, "\n")
	}
}

//type Individual struct {
//}

type LifeProblem struct {
	id         int
	start, end *Board_BoolPacked
	steps      int
	// Finished, iterations, confidence, etc
}

type LifeProblemSet struct {
	problem map[int]LifeProblem
}

func (s *LifeProblemSet) load_csv(f string, is_training bool, id_list []int) {
	if s.problem == nil {
		s.problem = make(map[int]LifeProblem)
	}
	file, err := os.Open(f)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// First line different
	header, err := reader.Read()
	if header[0] != "id" {
		fmt.Println("Bad Header", err)
		return
	}
	//fmt.Println("Header Start: ", header[2:402])
	//fmt.Println("Header Stop : ", header[402:802])

	id_max := 0
	id_map := make(map[int]bool)
	for _, id := range id_list {
		id_map[id] = true
		if id_max < id {
			id_max = id
		}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// record is []string
		id, _ := strconv.Atoi(record[0])
		if id_map[id] {
			//fmt.Println(record) // record has the type []string
			steps, _ := strconv.Atoi(record[1])

			start := NewBoard_BoolPacked(20, 20)
			end := NewBoard_BoolPacked(20, 20)
			if is_training {
				start.LoadArray(record[2:402])
				end.LoadArray(record[402:802])
			} else {
				end.LoadArray(record[2:402])
			}

			s.problem[id] = LifeProblem{
				id:    id,
				start: start,
				end:   end,
				steps: steps,
			}
			fmt.Printf("Loaded problem[%d] : steps=%d\n", id, steps)
			//fmt.Print(s.problem[id].start)
		}
		if id > id_max {
			return // fact-of-life : ids are ascending order
		}
	}
}

type ImageSet struct {
	im                       *image.RGBA
	rows, cols               int
	row_current, col_current int
}

func NewImageSet(rows, cols int) *ImageSet {
	im := image.NewRGBA(image.Rect(0, 0, cols*(20+2)+2, rows*(20+2)+2))                             //*NRGBA (image.Image interface)
	draw.Draw(im, im.Bounds(), image.NewUniform(color.RGBA{98, 166, 255, 255}), image.ZP, draw.Src) // color.Transparent
	return &ImageSet{
		im:   im,
		rows: rows, cols: cols,
		row_current: 0, col_current: 0,
	}
}

func (i *ImageSet) save(f string) {
	w, _ := os.Create(f)
	defer w.Close()
	png.Encode(w, i.im)
}

func (i *ImageSet) DrawStats(row, col int, bs *BoardStats) {
	offset_x := col*(20+2) + 2
	offset_y := row*(20+2) + 2

	for x := 0; x < bs.w; x++ {
		for y := 0; y < bs.h; y++ {
			g := uint8(bs.freq[x][y] * 255 / bs.count)
			i.im.Set(offset_x+x, offset_y+y, color.Gray{g})
		}
	}
}

func (i *ImageSet) DrawStatsNext(bs *BoardStats) {
	i.DrawStats(i.row_current, i.col_current, bs)
	i.col_current++
	if i.col_current >= i.cols {
		i.DrawStatsCRLF()
	}
}

func (i *ImageSet) DrawStatsCRLF() {
	i.col_current = 0
	i.row_current++
	if i.row_current >= i.rows {
		//fmt.Print("New Beginning\n")
		i.row_current = 0
	}
}

func main_verify_training_examples() {
	var kaggle LifeProblemSet

	problem_offset := 100

	id_list := []int{}
	id_map := make(map[int]bool)
	for id := problem_offset; id < problem_offset+10; id++ {
		id_list = append(id_list, id)
		id_map[id] = true
	}
	kaggle.load_csv("data/train.csv", true, id_list)
	//fmt.Println(kaggle.problem[107].start)
	//fmt.Println(kaggle.problem[107].end)

	image := NewImageSet(10, 11) // 10rows of 11 images, formatted 'appropriately'

	for _, id := range id_list {
		bs_start := NewBoardStats(20, 20)
		kaggle.problem[id].start.AddToStats(bs_start)

		bs_end := NewBoardStats(20, 20)
		kaggle.problem[id].end.AddToStats(bs_end)

		//image.DrawStats(r,c, bs)
		image.DrawStatsNext(bs_start)
		image.DrawStats(image.row_current, image.cols-1, bs_end)

		l := NewBoardIterator(20, 20)
		// NB: This destroys 'current' :: should have a copy function
		l.current = kaggle.problem[id].start

		for i := 0; i<kaggle.problem[id].steps; i++ {
			l.Iterate(1) // Just 1 step per image for now

			bs_now := NewBoardStats(20, 20)
			l.current.AddToStats(bs_now)
			image.DrawStatsNext(bs_now)
		}

		image.DrawStatsCRLF()
	}

	image.save("images/main.png")
}

func main() {
	main_verify_training_examples()
}
