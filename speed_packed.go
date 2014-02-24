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

func (attempt *Board_BoolPacked) CompareTo(target *Board_BoolPacked) int { // OPTIMIZED FOR BoolPacked
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
	}
	return r
}


/****************************************************************************************/
// The following will work on more generalized GoL mechanics, but are 10x slower
/****************************************************************************************/


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

// String returns the game board as a string.
func (f *Board_BoolPacked) String() string {
	var buf bytes.Buffer
	outer := 1
	for y := 0 - outer; y < f.h+outer; y++ {
		for x := 0 - outer; x < f.w+outer; x++ {
			b := byte('-')
			if x < 0 || x >= f.w || y < 0 || y >= f.h {
				b = '0'
			} else { 
				if f.isSet(x, y) {
					b = '*'
				}
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


type BoardStats struct {
	freq  [][]int
	w, h  int
	count int
	mismatch_amount int
}

// NewField_BoolArray returns an empty field of the specified width and height.
func NewBoardStats(w, h int) *BoardStats {
	freq := make([][]int, h)
	for i := range freq {
		freq[i] = make([]int, w)
	}
	//fmt.Print("CreatedBoardStats\n")
	return &BoardStats{freq: freq, w: w, h: h, count: 0, mismatch_amount:0}
}

func (bs *BoardStats) MisMatchBy(mismatch int) {
	bs.mismatch_amount = mismatch
}

// BoardIterator stores the state of a round of Conway's Game of Life.
type BoardIterator struct {
	current, temp_internal_only *Board_BoolPacked
}

// BoardIterator returns a new Life game state
func NewBoardIterator(w, h int) *BoardIterator {
	return &BoardIterator{
		current: NewBoard_BoolPacked(w, h), 
		temp_internal_only: NewBoard_BoolPacked(w, h),
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

func init() {
	fmt.Print("init() called\n")
	build_count_bits_array()
}

func main_orig() {
	l := NewBoardIterator(board_width, board_height)
	for i := 0; i < 65; i++ {
		l.Iterate(1)
		fmt.Print("\x0c", l) // Clear screen and print field.
		time.Sleep(time.Second / 30)
	}
}

func main_test_random() {
	for pct := float32(0.1); pct < 1.0; pct += 0.1 {
		l := NewBoardIterator(board_width, board_height)
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
		l := NewBoardIterator(board_width, board_height)
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
		l := NewBoardIterator(board_width, board_height)
		l.current.UniformRandom(pct + 0.01)

		//fmt.Print("\x0c", l)
		fmt.Print(l.current, "\n")
		l.Iterate(5)
		fmt.Print(l.current, "\n")
	}
}

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

			start := NewBoard_BoolPacked(board_width, board_height)
			end := NewBoard_BoolPacked(board_width, board_height)
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
			//fmt.Printf("Loaded problem[%d] : steps=%d\n", id, steps)
			//fmt.Print(s.problem[id].start)
		}
		if id > id_max {
			return // fact-of-life : ids are ascending order, so can quit reading early
		}
	}
}

type ImageSet struct {
	im                       *image.RGBA
	rows, cols               int
	row_current, col_current int
}

func NewImageSet(rows, cols int) *ImageSet {
	im := image.NewRGBA(image.Rect(0, 0, cols*(board_width+2)+2, rows*(board_height+2)+2))                             //*NRGBA (image.Image interface)
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
	offset_x := col*(board_width+2) + 2
	offset_y := row*(board_height+2) + 2

	for x := 0; x < bs.w; x++ {
		for y := 0; y < bs.h; y++ {
			g := bs.freq[x][y] * 255 / bs.count
			if bs.mismatch_amount>0 {
				pct := 100 - bs.mismatch_amount * 50 / 100
				if pct<0 {
					pct=0
				}
				//fmt.Printf("Mismatch pct=%d\n", pct)
				g = (g*pct) /100
			}
			i.im.Set(offset_x+x, offset_y+y, color.Gray{uint8(g)})
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
	for id := problem_offset; id < problem_offset+10; id++ {
		id_list = append(id_list, id)
	}
	kaggle.load_csv("data/train.csv", true, id_list)
	//fmt.Println(kaggle.problem[107].start)
	//fmt.Println(kaggle.problem[107].end)

	image := NewImageSet(10, 11) // 10 rows of 11 images each, formatted 'appropriately'

	for _, id := range id_list {
		bs_start := NewBoardStats(board_width, board_height)
		kaggle.problem[id].start.AddToStats(bs_start)

		bs_end := NewBoardStats(board_width, board_height)
		kaggle.problem[id].end.AddToStats(bs_end)

		//image.DrawStats(r,c, bs)
		image.DrawStatsNext(bs_start)
		image.DrawStats(image.row_current, image.cols-1, bs_end)

		l := NewBoardIterator(board_width, board_height)
		l.current.CopyFrom(kaggle.problem[id].start)

		for i := 0; i < kaggle.problem[id].steps; i++ {
			l.Iterate(1) // Just 1 step per image for now

			bs_now := NewBoardStats(board_width, board_height)
			l.current.AddToStats(bs_now)
			image.DrawStatsNext(bs_now)
		}
		image.DrawStatsNext(bs_end) // For ease of comparison..
	
		if mismatch := kaggle.problem[id].end.CompareTo(l.current); mismatch>0 {
			fmt.Printf("Training Example[%d] FAIL - by %d\n", id, mismatch)
		}

		image.DrawStatsCRLF()
	}

	image.save("images/train.png")
}

func main_visualize_density() {
	image := NewImageSet(10, 11) // 10 rows of 11 images each, formatted 'appropriately'

	for pct:=float32(0.1); pct<0.99; pct+=0.1 {
		var bs []*BoardStats
		for j:=0; j<10; j++ {
			bs = append(bs, NewBoardStats(board_width, board_height))
		}

		for i:=0; i<1000; i++ {
			l := NewBoardIterator(board_width, board_height)
			l.current.UniformRandom(pct)
			l.current.AddToStats(bs[0])
			for j:=1; j<len(bs); j++ {
				l.Iterate(1)
				l.current.AddToStats(bs[j])
			}
		}
		for j:=0; j<len(bs); j++ {
			image.DrawStatsNext(bs[j])
		}
		image.DrawStatsCRLF()
	}

	image.save("images/density.png")
}

type Individual struct {
	start *Board_BoolPacked
	//end *Board_BoolPacked
	fitness int  // higher is better, no particular scale
}

/*
func IndividualMutationGenerator(amount int) func f(i *Individual) *Individual{
	return func(i *Individual) *Individual{
		
	}
}
*/

type Population struct {
	individual []*Individual
	
	pressure_pct int
	
	mutation_pct int // (0..100)
	mutation_radius int
	mutation_loop_pct int

	crossover_pct int // (0..100)
}

func NewPopulation(size int, radius int) *Population {
	//fmt.Printf("NewPopulation(size=%d)\n", size)
	ind := make([]*Individual, size)
	for i:=0; i<size; i++ {
		ind[i] = &Individual{ start:NewBoard_BoolPacked(board_width, board_height), fitness:0 }
	}
	//fmt.Printf("NewPopulation(size=%d) inited\n", size)
	return &Population{
		individual:ind,
		
		pressure_pct:90,
		
		mutation_pct:30,
		mutation_radius:radius,
		mutation_loop_pct:20,
		
		crossover_pct:30*0,
	}
}

func (p *Population) PickIndividualWithPressure() *Individual {  // pressure_pct is in (50..100)
	// Pick two individuals at random from population
	i_1 := p.individual[rand.Intn(len(p.individual))]
	i_2 := p.individual[rand.Intn(len(p.individual))]
	
	if i_1.fitness < i_2.fitness {
		i_1,i_2 = i_2,i_1 // Switch them so that i_1 is the fitter (higher is better) of the two
	}
	
	// if pct< a threshold, pick the better one
	i_chosen := i_1
	if rand.Intn(100) > p.pressure_pct { // i.e. only sometimes do the opposite
		i_chosen = i_2
	}
	return i_chosen
}

func (p *Population) GenerationAfter(prev *Population) {
	// Fill in every slot
	for c:=0; c<len(p.individual); c++ {
		//fmt.Printf("Fitness to choose : {%d,%d} -> %d\n", i_1.fitness, i_2.fitness, i_chosen.fitness)
		
		choser := rand.Intn(100)
		if 0<=choser && choser < p.crossover_pct { 
			// Do a 'crossover copy' from two individuals in previous population to this one
			parent_1 := prev.PickIndividualWithPressure()
			parent_2 := prev.PickIndividualWithPressure()
			p.individual[c].start.CrossoverFrom(parent_1.start, parent_2.start)
		} else { // Do a simple copy, with the possibility of mutation (below)
			i_chosen := prev.PickIndividualWithPressure()
			p.individual[c].start.CopyFrom(i_chosen.start)
			if p.crossover_pct<=choser && choser < (p.crossover_pct + p.mutation_pct) {
				p.individual[c].start.MutateRadiusBits(p.mutation_loop_pct, p.mutation_radius) // % do additional mutation, radius of action
			}
		}

		p.individual[c].fitness = 0
	}
}

/*
func (p *Population) MutateIndividuals(another_mutation_pct int, radius int) {  // % do additional mutation, radius of action
	for c:=0; c<len(p.individual); c++ {
		//p.individual[c].start.MutateFlipBits(rand.Intn(mutation_size))
		p.individual[c].start.MutateRadiusBits(another_mutation_pct, radius)
	}
}
*/

func main_population_score() {
	image := NewImageSet(10, 12) // 10 rows of 12 images each, formatted 'appropriately'
	
	var kaggle LifeProblemSet
	id := 107
	kaggle.load_csv("data/train.csv", true, []int{id}) 

	problem := kaggle.problem[id]

	// This is the TRUE starting place : for reference
	bs_start := NewBoardStats(board_width, board_height)
	problem.start.AddToStats(bs_start)

	// This is the TRUE ending place : for reference
	bs_end := NewBoardStats(board_width, board_height)
	problem.end.AddToStats(bs_end)

	// Create a population of potential boards
	pop_size := 5
	pop := NewPopulation(pop_size, problem.steps)
	for i:=0; i<pop_size; i++ {
		// Create a candidate starting point
		// NB:  We can only work from the problem.end
		pop.individual[i].start.CopyFrom(problem.end)
		//pop.individual[i].start.UniformRandom(0.4)
	}
	
	p_temp := NewPopulation(pop_size, problem.steps)

	l := NewBoardIterator(board_width, board_height)
	
	iter_max := 10
	for iter:=0; iter<iter_max; iter++ {
		disp_row := (0 == (iter) % (iter_max/10))
		
		if disp_row {
			// for ease of comparison
			image.DrawStatsNext(bs_start)
		}
		
		// Evaluate fitness of every individual in pop
		//for i:=0; i<pop_size; i++ {
		for i, individual := range pop.individual {
			l.current.CopyFrom(individual.start)
			
			//l.Iterate(5)
			
			if i<5 && disp_row {
				bs_trial := NewBoardStats(board_width, board_height)
				l.current.AddToStats(bs_trial)
				
				mismatch_from_true_start := problem.start.CompareTo(l.current)
				fmt.Printf("\n%3d.%2d : Mismatch from true start = %d\n", iter, i, mismatch_from_true_start)
				bs_trial.MisMatchBy(mismatch_from_true_start)
			
				image.DrawStatsNext(bs_trial)
			}
			
			l.Iterate(problem.steps)
			
			mismatch_from_true_end := problem.end.CompareTo(l.current)
			individual.fitness = -mismatch_from_true_end
			
			if i<5 && disp_row {
				bs_result := NewBoardStats(board_width, board_height)
				l.current.AddToStats(bs_result)
				
				fmt.Printf("%3d.%2d : Mismatch from true end   = %d\n", iter, i, mismatch_from_true_end)
				bs_result.MisMatchBy(mismatch_from_true_end)
				
				image.DrawStatsNext(bs_result)
			}
		}
		
		if disp_row {
			//image.DrawStatsNext(bs_end) // For ease of comparison..
			image.DrawStats(image.row_current, image.cols-1, bs_end)
			
			image.DrawStatsCRLF()
		}
		
		p_temp.GenerationAfter(pop)
		pop, p_temp = p_temp, pop // Switcheroo to advance to next population
	}

	//image.DrawStats(image.row_current, image.cols-1, bs_end)
	
	image.save("images/score_mutated.png")
}

func main() {
	//main_timer()
	//main_verify_training_examples()
	//main_visualize_density()
	main_population_score()
}
